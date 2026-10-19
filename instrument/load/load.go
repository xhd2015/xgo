package load

import (
	"bytes"
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
	Test        bool   // -test flag, test variants are mapped back to plain packages
	// GoVersion of the go command, lets go list skip unused fields
	GoVersion *goinfo.GoVersion

	// OverlayFile is passed to `go list -overlay=` during package discovery.
	// Overlay (in-memory) is used only when reading/parsing file contents.
	// Callers that rely on go.mod replacements visible only through an overlay
	// (e.g. doctest vendor-gomod bridges) must set OverlayFile or discovery
	// returns Incomplete packages with empty GoFiles.
	OverlayFile string

	// max file size to parse
	// if file size is larger than this
	// see https://github.com/xhd2015/xgo/issues/303
	// for more background
	MaxFileSize int64

	Goroot string

	FilterErrorFile bool

	Fset *token.FileSet

	// ParseFilter, if set, selects the packages whose files are read and
	// parsed. The others are returned with Deferred set and no Files, and
	// can be parsed later with ParseDeferred.
	ParseFilter func(pkg *goinfo.Package) bool

	// OnList, if set, sees every listed package before ParseFilter is
	// called, so the filter can depend on the whole listing.
	OnList func(pkgs []*goinfo.Package)
	// Listed, if set, returns the result of listing the packages, as
	// ListOptions describes, which the caller started earlier
	Listed func() ([]*goinfo.Package, error)

	// TestFilter, if set, selects the packages whose test files are parsed
	// when IncludeTest is set.
	TestFilter func(pkg *goinfo.Package) bool
}

type Package struct {
	GoPackage *goinfo.Package
	Files     []*File

	// Deferred reports that Files were not loaded yet, see
	// LoadOptions.ParseFilter.
	Deferred bool
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

// ListOptions are the options LoadPackages lists packages with.
func ListOptions(opts LoadOptions) goinfo.LoadPackageOptions {
	return goinfo.LoadPackageOptions{
		Dir:         opts.Dir,
		Mod:         opts.Mod,
		ModFile:     opts.ModFile,
		Goroot:      opts.Goroot,
		Deps:        opts.Deps,
		Test:        opts.Test,
		OverlayFile: opts.OverlayFile,
		GoVersion:   opts.GoVersion,
	}
}

func LoadPackages(args []string, opts LoadOptions) (*Packages, error) {
	fset := opts.Fset

	var pkgs []*goinfo.Package
	var err error
	if opts.Listed != nil {
		pkgs, err = opts.Listed()
	} else {
		pkgs, err = goinfo.ListPackages(args, ListOptions(opts))
	}
	if err != nil {
		return nil, err
	}

	if opts.OnList != nil {
		opts.OnList(pkgs)
	}
	if fset == nil {
		fset = token.NewFileSet()
	}
	loadPkgs := make([]*Package, len(pkgs))
	var parsePkgs []*Package
	for i, pkg := range pkgs {
		loadPkgs[i] = &Package{
			GoPackage: pkg,
		}
		if opts.ParseFilter != nil && !opts.ParseFilter(pkg) {
			loadPkgs[i].Deferred = true
			continue
		}
		parsePkgs = append(parsePkgs, loadPkgs[i])
	}
	parsePackageFiles(parsePkgs, opts, fset)

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
	// Normalize CRLF to LF: git core.autocrlf on Windows injects \r
	// into all files including raw string literals inside backticks.
	// The Go scanner strips \r from BasicLit.Value but token.FileSet
	// byte offsets include them, causing BasicLit.End() and thus
	// GenDecl.End() to be short by the \r count. Normalizing before
	// parsing eliminates this discrepancy entirely.
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	return content, err
}

func parseFile(fset *token.FileSet, absFilePath string, content []byte) (*ast.File, error) {
	return parser.ParseFile(fset, string(absFilePath), content, parser.ParseComments|parser.SkipObjectResolution)
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

	file, err := parser.ParseFile(fset, string(absFilePath), content, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		f.Error = err
		return f
	}

	f.Syntax = file
	return f
}

// ParseDeferred reads and parses the files of deferred packages, in place.
func ParseDeferred(pkgs []*Package, opts LoadOptions) {
	var parsePkgs []*Package
	for _, pkg := range pkgs {
		if pkg.Deferred {
			pkg.Deferred = false
			parsePkgs = append(parsePkgs, pkg)
		}
	}
	fset := opts.Fset
	if fset == nil {
		fset = token.NewFileSet()
	}
	parsePackageFiles(parsePkgs, opts, fset)
}

func parsePackageFiles(loadPkgs []*Package, opts LoadOptions, fset *token.FileSet) {
	if len(loadPkgs) == 0 {
		return
	}
	overlayFS := opts.Overlay
	maxFileSize := opts.MaxFileSize
	filterErrorFile := opts.FilterErrorFile

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
		if opts.IncludeTest && (opts.TestFilter == nil || opts.TestFilter(pkg.GoPackage)) {
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
}
