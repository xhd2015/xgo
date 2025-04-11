package resolve

import (
	"go/ast"
	"go/token"

	astutil "github.com/xhd2015/xgo/instrument/ast"
	"github.com/xhd2015/xgo/instrument/edit"
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
		fileDecls := file.Decls
		for _, decl := range file.File.Syntax.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				case token.VAR:
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
								fileDecls = append(fileDecls, &edit.Decl{
									Kind:  edit.DeclKindVar,
									Ident: name,
									Type:  spec.Type,
									Value: value,
									Decl:  decl,
									File:  file,
								})
							}
						}
					}
				case token.CONST:
					// TODO
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
func Traverse(registry PackageRegistry, packages []*edit.Package, recorder *Recorder) error {
	// rewrite selectors
	global := &GlobalScope{
		Packages:      registry,
		Recorder:      recorder,
		detectVarTrap: true,
		detectMock:    true,
	}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			traverseFuncDecls(global, pkg, file)
		}
	}

	return nil
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

func (c *packagesRegistry) GetPackage(pkgPath string) *edit.Package {
	return c.packages.PackageByPath[pkgPath]
}

func NewPackagesRegistry(packages *edit.Packages) PackageRegistry {
	return &packagesRegistry{
		packages: packages,
	}
}

func traverseFuncDecls(global *GlobalScope, pkg *edit.Package, file *edit.File) {
	fileScope := newFileScope(global, pkg, file)
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
		fileScope.traverseFunc(fnDecl.Recv, fnDecl.Type, fnDecl.Body)
	}
}
