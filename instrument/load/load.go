package load

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/overlay"
	"github.com/xhd2015/xgo/support/goinfo"
)

type LoadOptions struct {
	Dir         string
	Overlay     overlay.Overlay
	Mod         string
	IncludeTest bool
	ModFile     string // -modfile flag
	Deps        bool   // -deps flag

	// max file size to parse
	// if file size is larger than this
	// see https://github.com/xhd2015/xgo/issues/303
	// for more background
	MaxFileSize int64

	Goroot string

	FilterErrorFile bool

	Fset *token.FileSet
}

type Package struct {
	GoPackage *goinfo.Package
	Files     []*File
}

type File struct {
	AbsPath string
	Name    string
	Content []byte
	Error   error
	Syntax  *ast.File
}

type Packages struct {
	Fset     *token.FileSet
	Packages []*Package
}

func LoadPackages(args []string, opts LoadOptions) (*Packages, error) {
	dir := opts.Dir
	overlayFS := opts.Overlay
	deps := opts.Deps
	mod := opts.Mod
	modFile := opts.ModFile
	maxFileSize := opts.MaxFileSize
	filterErrorFile := opts.FilterErrorFile
	goroot := opts.Goroot
	fset := opts.Fset

	pkgs, err := goinfo.ListPackages(args, goinfo.LoadPackageOptions{
		Dir:     dir,
		Mod:     mod,
		ModFile: modFile,
		Goroot:  goroot,
		Deps:    deps,
		Test:    false, // NOTE: don't set it
	})
	if err != nil {
		return nil, err
	}

	if fset == nil {
		fset = token.NewFileSet()
	}
	loadPkgs := make([]*Package, len(pkgs))
	for i, pkg := range pkgs {
		loadPkgs[i] = &Package{
			GoPackage: pkg,
		}
	}

	begin := time.Now()

	var wg sync.WaitGroup
	done := make(chan struct{})

	// credit to https://github.com/golang/tools/blob/4ec26d68b3c042c274fa5dcc633cb014846e2dd9/go/packages/packages.go#L1332
	// see https://github.com/xhd2015/xgo/issues/336
	const IO_LIMIT = 20
	var CPU_LIMIT = runtime.GOMAXPROCS(0) // this could be 10

	readChan := make(chan *File, IO_LIMIT)
	parseChan := make(chan *File, CPU_LIMIT)

	// parallelize readers and parsers
	for i := 0; i < IO_LIMIT; i++ {
		go func() {
			for {
				select {
				case f := <-readChan:
					content, err := readFile(overlayFS, f.AbsPath, maxFileSize)
					if err != nil {
						// mark the file done
						wg.Done()
						f.Error = err
						continue
					}
					f.Content = content
					parseChan <- f
				case <-done:
					return
				}
			}
		}()
	}

	// Parsing is CPU intensive
	for i := 0; i < CPU_LIMIT; i++ {
		go func() {
			for {
				select {
				case f := <-parseChan:
					syntax, err := parseFile(fset, f.AbsPath, f.Content)

					f.Error = err
					f.Syntax = syntax

					// mark the file done
					// see https://github.com/xhd2015/xgo/issues/344
					wg.Done()
				case <-done:
					return
				}
			}
		}()
	}

	// parse packages
	for _, pkg := range loadPkgs {
		var files []string
		for _, file := range pkg.GoPackage.GoFiles {
			if !strings.HasSuffix(file, ".go") {
				continue
			}
			files = append(files, file)
		}
		if opts.IncludeTest {
			for _, file := range pkg.GoPackage.TestGoFiles {
				if !strings.HasSuffix(file, ".go") {
					continue
				}
				files = append(files, file)
			}
			for _, file := range pkg.GoPackage.XTestGoFiles {
				if !strings.HasSuffix(file, ".go") {
					continue
				}
				files = append(files, file)
			}
		}
		if len(files) == 0 {
			continue
		}
		pkgFiles := make([]*File, len(files))
		for i, file := range files {
			parsingFile := &File{
				AbsPath: filepath.Join(pkg.GoPackage.Dir, file),
				Name:    file,
			}
			pkgFiles[i] = parsingFile

			// enqueue a file
			wg.Add(1)
			readChan <- parsingFile
		}
		pkg.Files = pkgFiles
	}
	wg.Wait()
	close(done)

	// filter error files
	var numFiles int
	var numErrFiles int
	for _, pkg := range loadPkgs {
		n := len(pkg.Files)
		j := 0
		for i := 0; i < n; i++ {
			file := pkg.Files[i]
			if filterErrorFile && file.Error != nil {
				continue
			}
			pkg.Files[j] = file
			j++
		}
		pkg.Files = pkg.Files[:j]

		numFiles += n
		numErrFiles += n - j
	}

	// serial loading:
	// - LoadPackages: loaded 173 packages, parsed 1636 files, cost 1.054140333s
	// - LoadPackages: loaded 2250 packages, parsed 14565 files, cost 4.262132333s
	// parallel loading:
	// - LoadPackages: loaded 2487 packages, parsed 11973 files, 6 files error, cost 709.993875ms
	if config.EnabledLogDebug() {
		config.LogDebug("LoadPackages: loaded %d packages, parsed %d files, %d files error, cost %v", len(loadPkgs), numFiles, numErrFiles, time.Since(begin))
	}

	return &Packages{
		Fset:     fset,
		Packages: loadPkgs,
	}, nil
}

func readFile(overlayFS overlay.Overlay, absFilePath string, maxFileSize int64) ([]byte, error) {
	if maxFileSize > 0 {
		size, err := overlayFS.Size(overlay.AbsFile(absFilePath))
		if err != nil {
			return nil, err
		}
		if size > maxFileSize {
			return nil, fmt.Errorf("file size %d large than %d", size, maxFileSize)
		}
	}
	_, content, err := overlayFS.ReadBytes(overlay.AbsFile(absFilePath))
	return content, err
}

func parseFile(fset *token.FileSet, absFilePath string, content []byte) (*ast.File, error) {
	return parser.ParseFile(fset, string(absFilePath), content, parser.ParseComments)
}

func (c *Packages) Filter(f func(pkg *Package) bool) *Packages {
	var filtered []*Package
	for _, pkg := range c.Packages {
		if f(pkg) {
			filtered = append(filtered, pkg)
		}
	}
	return &Packages{
		Fset:     c.Fset,
		Packages: filtered,
	}
}

func readAndParseFile(fset *token.FileSet, absFilePath string, overlayFS overlay.Overlay) *File {
	f := &File{
		AbsPath: absFilePath,
		Name:    filepath.Base(absFilePath),
	}

	_, content, err := overlayFS.ReadBytes(overlay.AbsFile(absFilePath))
	if err != nil {
		f.Error = err
		return f
	}
	f.Content = content

	file, err := parser.ParseFile(fset, string(absFilePath), content, parser.ParseComments)
	if err != nil {
		f.Error = err
		return f
	}

	f.Syntax = file
	return f
}
