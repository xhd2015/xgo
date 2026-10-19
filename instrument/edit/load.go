package edit

import (
	"sort"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/load"
	"github.com/xhd2015/xgo/support/goinfo"
)

func (c *Packages) LoadPackage(pkgPath string) (*Package, bool, error) {
	pkg, ok := c.PackageByPath[pkgPath]
	if ok {
		return pkg, true, nil
	}
	if deferred, ok := c.deferred[pkgPath]; ok {
		c.parseDeferred([]*load.Package{deferred})
		return c.PackageByPath[pkgPath], true, nil
	}

	// usually triggered by test-only imports
	config.LogDebug("trigger load package: %s", pkgPath)
	pkgs, err := load.LoadPackages([]string{pkgPath}, c.LoadOptions)
	if err != nil {
		return nil, false, err
	}
	c.Add(pkgs)
	if deferred, ok := c.deferred[pkgPath]; ok {
		c.parseDeferred([]*load.Package{deferred})
	}
	return c.PackageByPath[pkgPath], false, nil
}

// PackageName returns the name of the package pkgPath if it is listed,
// whether parsed or deferred, or "" otherwise.
func (c *Packages) PackageName(pkgPath string) string {
	if pkg, ok := c.PackageByPath[pkgPath]; ok {
		return pkg.LoadPackage.GoPackage.Name
	}
	if pkg, ok := c.deferred[pkgPath]; ok {
		return pkg.GoPackage.Name
	}
	return ""
}

func (c *Packages) LoadPackages(pkgPaths []string) error {
	var missing []string
	var deferred []*load.Package
	for _, pkgPath := range pkgPaths {
		if _, ok := c.PackageByPath[pkgPath]; ok {
			continue
		}
		if pkg, ok := c.deferred[pkgPath]; ok {
			deferred = append(deferred, pkg)
			continue
		}
		missing = append(missing, pkgPath)
	}
	c.parseDeferred(deferred)
	if len(missing) == 0 {
		return nil
	}
	missingPkgs, err := load.LoadPackages(missing, c.LoadOptions)
	if err != nil {
		return err
	}
	c.Add(missingPkgs)
	return nil
}

// ListPackages lists pkgPaths that are neither parsed nor deferred yet, in
// one `go list`, without parsing them: LoadPackage parses them when needed.
func (c *Packages) ListPackages(pkgPaths []string) error {
	var missing []string
	seen := make(map[string]bool, len(pkgPaths))
	for _, pkgPath := range pkgPaths {
		if seen[pkgPath] {
			continue
		}
		seen[pkgPath] = true
		if _, ok := c.PackageByPath[pkgPath]; ok {
			continue
		}
		if _, ok := c.deferred[pkgPath]; ok {
			continue
		}
		missing = append(missing, pkgPath)
	}
	if len(missing) == 0 {
		return nil
	}
	opts := c.LoadOptions
	opts.ParseFilter = func(*goinfo.Package) bool { return false }
	pkgs, err := load.LoadPackages(missing, opts)
	if err != nil {
		return err
	}
	c.Add(pkgs)
	return nil
}

// ParseDeferred parses every deferred package and adds it to Packages.
func (c *Packages) ParseDeferred() {
	pkgs := make([]*load.Package, 0, len(c.deferred))
	for _, pkg := range c.deferred {
		pkgs = append(pkgs, pkg)
	}
	// deterministic order of Packages
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].GoPackage.ImportPath < pkgs[j].GoPackage.ImportPath
	})
	c.parseDeferred(pkgs)
}

// TakeDeferred removes the listed but not yet parsed package pkgPath and
// returns it, or nil if there is none. Parse it with ParseDeferredPackage
// and add it back with AddParsed: the parsing needs no access to c, so it
// can run while c is used for other packages.
func (c *Packages) TakeDeferred(pkgPath string) *load.Package {
	pkg, ok := c.deferred[pkgPath]
	if !ok {
		return nil
	}
	delete(c.deferred, pkgPath)
	return pkg
}

// ParseDeferredPackage parses the files of pkg, taken by TakeDeferred.
func (c *Packages) ParseDeferredPackage(pkg *load.Package) {
	opts := c.LoadOptions
	// positions must belong to the FileSet the packages are registered with
	opts.Fset = c.Fset
	load.ParseDeferred([]*load.Package{pkg}, opts)
}

// AddParsed adds pkg, parsed by ParseDeferredPackage.
func (c *Packages) AddParsed(pkg *load.Package) {
	c.Add(&load.Packages{Fset: c.Fset, Packages: []*load.Package{pkg}})
}

func (c *Packages) parseDeferred(pkgs []*load.Package) {
	if len(pkgs) == 0 {
		return
	}
	opts := c.LoadOptions
	// positions must belong to the FileSet the packages are registered with
	opts.Fset = c.Fset
	load.ParseDeferred(pkgs, opts)
	c.Add(&load.Packages{Fset: c.Fset, Packages: pkgs})
}

func InitPkgFlag(pkg *Package) {
	isXgo, allow := config.CheckInstrument(pkg.LoadPackage.GoPackage.ImportPath)
	if isXgo {
		pkg.Xgo = true
	}
	if allow {
		pkg.AllowInstrument = true
	}
}
