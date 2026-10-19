package resolve

import (
	"go/ast"
	"go/token"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	astutil "github.com/xhd2015/xgo/instrument/ast"
	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/config/config_debug"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/load"
)

func CollectDecls(pkg *edit.Package) {
	// first, collect all toplevel variables
	if pkg.Collected {
		return
	}
	decls := pkg.Decls
	if decls == nil {
		decls = make(map[string]*edit.Decl)
	}
	typesMethods := make(map[string]map[string]*edit.FuncDecl)
	for _, file := range pkg.Files {
		if config.DEBUG {
			config_debug.OnCollectFileDecl(pkg, file)
		}
		fileDecls := file.Decls
		for _, decl := range file.File.Syntax.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				case token.VAR, token.CONST:
					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						case *ast.ValueSpec:
							for i, name := range spec.Names {
								if isBlankName(name.Name) {
									continue
								}
								var value ast.Expr
								if i < len(spec.Values) {
									value = spec.Values[i]
								}
								kind := edit.DeclKindVar
								if decl.Tok == token.CONST {
									kind = edit.DeclKindConst
								}
								fileDecls = append(fileDecls, &edit.Decl{
									Kind:  kind,
									Ident: name,
									Type:  spec.Type,
									Value: value,
									Decl:  decl,
									File:  file,
								})
							}
						}
					}
				case token.TYPE:
					// type Some...
					for _, spec := range decl.Specs {
						spec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						if isBlankName(spec.Name.Name) {
							continue
						}
						// TODO: check spec.Assign to handle
						// alias if necessary
						fileDecls = append(fileDecls, &edit.Decl{
							Kind:  edit.DeclKindType,
							Ident: spec.Name,
							Type:  spec.Type,
							Decl:  decl,
							File:  file,
						})
					}
				}
			case *ast.FuncDecl:
				funcName := decl.Name.Name
				if isBlankName(funcName) {
					continue
				}
				fnDecl := &edit.Decl{
					Kind:  edit.DeclKindFunc,
					Ident: decl.Name,
					FuncDecl: &edit.FuncDecl{
						Name:   funcName,
						Syntax: decl,
					},
					File: file,
				}
				if decl.Recv != nil && len(decl.Recv.List) > 0 {
					recv := decl.Recv.List[0]
					if recv.Type != nil {
						ptr, _, typIdt := astutil.ParseReceiverType(recv.Type)
						fnDecl.FuncDecl.RecvPtr = ptr
						typeName := typIdt.Name
						typeMethods := typesMethods[typeName]
						if typeMethods == nil {
							typeMethods = make(map[string]*edit.FuncDecl, 1)
							typesMethods[typeName] = typeMethods
						}
						typeMethods[funcName] = fnDecl.FuncDecl
						continue
					}
				} else {
					fileDecls = append(fileDecls, fnDecl)
				}
			}
		}
		for _, decl := range fileDecls {
			decls[decl.Ident.Name] = decl
		}
		file.Decls = fileDecls
	}
	for typeName, methods := range typesMethods {
		decl := decls[typeName]
		if decl == nil || decl.Kind != edit.DeclKindType {
			continue
		}
		decl.Methods = methods
	}
	pkg.Decls = decls
	pkg.Collected = true
}

// Traverse applies variable rewriting, and find possible
// functions `fn` such that `mock.Patch(fn)` is called.
//
// Files are traversed in parallel. What traversing a file changes outside
// of it (references recorded on declarations, the recorder) is buffered per
// file and applied in file order afterwards, so the result does not depend
// on the scheduling.
func Traverse(registry PackageRegistry, packages []*edit.Package, recorder *Recorder) error {
	type fileItem struct {
		pkg  *edit.Package
		file *edit.File
	}
	var items []fileItem
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			if !pkg.Main && strings.HasSuffix(file.File.Name, "_test.go") {
				// skip test files outside main package
				continue
			}
			items = append(items, fileItem{pkg: pkg, file: file})
		}
	}
	effects := make([]fileEffects, len(items))
	reg := &syncRegistry{registry: registry}

	workers := runtime.GOMAXPROCS(0)
	if config.DEBUG {
		// the debug hooks expect one file at a time
		workers = 1
	} else if workers > len(items) {
		workers = len(items)
	}
	var next int64 = -1
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			// the caches of a scope are per worker
			global := &GlobalScope{
				Packages:      reg,
				detectVarTrap: true,
				detectMock:    true,
			}
			for {
				i := int(atomic.AddInt64(&next, 1))
				if i >= len(items) {
					return
				}
				item := items[i]
				if config.DEBUG {
					config_debug.OnTraverseFile(item.pkg, item.file)
				}
				// Expressions are resolved, and looked up again, within the
				// file being traversed. Keeping the results of every file grew
				// the cache to every expression of the program, and let a
				// result derived from an unfinished (cyclic) resolution in one
				// file shadow the actual type in all files after it.
				global.ExprInfo = nil
				global.Recorder = &effects[i].recorder
				global.effects = &effects[i]
				traverseFuncDecls(global, item.pkg, item.file)
			}
		}()
	}
	wg.Wait()

	for i := range effects {
		for _, ref := range effects[i].varRefs {
			ref.decl.VarRefs = append(ref.decl.VarRefs, ref.ref)
		}
		recorder.merge(&effects[i].recorder)
	}
	return nil
}

// fileEffects is what traversing one file changes outside of it.
type fileEffects struct {
	varRefs  []declVarRef
	recorder Recorder
}

type declVarRef struct {
	decl *edit.Decl
	ref  *edit.VarRef
}

// syncRegistry makes a registry safe for the parallel traversal. A package
// is handed out only with its declarations collected, so that the workers
// can read them without locking. Packages listed but not parsed yet are
// parsed outside of the lock, so that the workers parse different packages
// in parallel; a package being parsed is waited for, not reported missing,
// so the lookups give the same answers whatever the scheduling.
type syncRegistry struct {
	mu       sync.Mutex
	registry PackageRegistry
	inflight map[string]*inflightPackage
}

type inflightPackage struct {
	pkg  *load.Package
	done chan struct{}
}

// deferredRegistry is implemented by registries that can parse a listed
// package without being locked.
type deferredRegistry interface {
	TakeDeferred(pkgPath string) *load.Package
	ParseDeferredPackage(pkg *load.Package)
	AddParsed(pkg *load.Package)
}

func (c *syncRegistry) Fset() *token.FileSet {
	// token.FileSet is safe for concurrent use
	return c.registry.Fset()
}

func (c *syncRegistry) LoadPackage(pkgPath string) (*edit.Package, bool, error) {
	c.mu.Lock()
	if f := c.inflight[pkgPath]; f != nil {
		c.mu.Unlock()
		<-f.done
		c.mu.Lock()
	}
	if pkg := c.registry.GetPackage(pkgPath); pkg != nil {
		CollectDecls(pkg)
		c.mu.Unlock()
		return pkg, true, nil
	}
	dr, ok := c.registry.(deferredRegistry)
	var deferred *load.Package
	if ok {
		deferred = dr.TakeDeferred(pkgPath)
	}
	if deferred == nil {
		// not listed: go list it, rare enough to do under the lock
		defer c.mu.Unlock()
		pkg, ok, err := c.registry.LoadPackage(pkgPath)
		if pkg != nil {
			CollectDecls(pkg)
		}
		return pkg, ok, err
	}
	f := &inflightPackage{pkg: deferred, done: make(chan struct{})}
	if c.inflight == nil {
		c.inflight = make(map[string]*inflightPackage)
	}
	c.inflight[pkgPath] = f
	c.mu.Unlock()

	dr.ParseDeferredPackage(deferred)

	c.mu.Lock()
	dr.AddParsed(deferred)
	pkg := c.registry.GetPackage(pkgPath)
	if pkg != nil {
		CollectDecls(pkg)
	}
	delete(c.inflight, pkgPath)
	close(f.done)
	c.mu.Unlock()
	return pkg, true, nil
}

func (c *syncRegistry) GetPackage(pkgPath string) *edit.Package {
	c.mu.Lock()
	if f := c.inflight[pkgPath]; f != nil {
		c.mu.Unlock()
		<-f.done
		c.mu.Lock()
	}
	defer c.mu.Unlock()
	pkg := c.registry.GetPackage(pkgPath)
	if pkg != nil {
		CollectDecls(pkg)
	}
	return pkg
}

func (c *syncRegistry) PackageName(pkgPath string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if f := c.inflight[pkgPath]; f != nil {
		return f.pkg.GoPackage.Name
	}
	return c.registry.PackageName(pkgPath)
}

type packagesRegistry struct {
	packages *edit.Packages
}

func (c *packagesRegistry) Fset() *token.FileSet {
	return c.packages.Fset
}

func (c *packagesRegistry) LoadPackage(pkgPath string) (*edit.Package, bool, error) {
	return c.packages.LoadPackage(pkgPath)
}

func (c *packagesRegistry) TakeDeferred(pkgPath string) *load.Package {
	return c.packages.TakeDeferred(pkgPath)
}

func (c *packagesRegistry) ParseDeferredPackage(pkg *load.Package) {
	c.packages.ParseDeferredPackage(pkg)
}

func (c *packagesRegistry) AddParsed(pkg *load.Package) {
	c.packages.AddParsed(pkg)
}

func (c *packagesRegistry) PackageName(pkgPath string) string {
	return c.packages.PackageName(pkgPath)
}

func (c *packagesRegistry) GetPackage(pkgPath string) *edit.Package {
	return c.packages.PackageByPath[pkgPath]
}

func NewPackagesRegistry(packages *edit.Packages) PackageRegistry {
	return &packagesRegistry{
		packages: packages,
	}
}

func traverseFuncDecls(global *GlobalScope, pkg *edit.Package, file *edit.File) {
	fileScope := newFileScope(global, pkg, file, 0)
	// NOTE: statements like this will be ignored:
	//  var _ = func() bool {
	//     x = true()
	//    return true
	// }
	for _, decl := range file.File.Syntax.Decls {
		fnDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fnDecl.Body == nil {
			continue
		}
		if config.DEBUG {
			config_debug.OnTraverseFuncDecl(pkg, file, fnDecl)
		}
		fileScope.traverseFunc(fnDecl.Recv, fnDecl.Name.Name, fnDecl.Type, fnDecl.Body)
	}
}
