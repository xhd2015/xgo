package edit

import (
	"github.com/xhd2015/xgo/instrument/load"
)

func (c *Packages) LoadPackage(pkgPath string) (*Package, bool, error) {
	pkg, ok := c.PackageByPath[pkgPath]
	if ok {
		return pkg, true, nil
	}

	pkgs, err := load.LoadPackages([]string{pkgPath}, c.LoadOptions)
	if err != nil {
		return nil, false, err
	}
	c.Add(pkgs)
	return c.PackageByPath[pkgPath], false, nil
}

func (c *Packages) LoadPackages(pkgPaths []string) error {
	var missing []string
	for _, pkgPath := range pkgPaths {
		_, ok := c.PackageByPath[pkgPath]
		if !ok {
			missing = append(missing, pkgPath)
		}
	}
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
