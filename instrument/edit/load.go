package edit

import (
	"github.com/xhd2015/xgo/instrument/load"
)

func (c *Packages) Load(pkgPath string) (*Package, error) {
	pkg, ok := c.PackageByPath[pkgPath]
	if ok {
		return pkg, nil
	}

	pkgs, err := load.LoadPackages([]string{pkgPath}, c.LoadOptions)
	if err != nil {
		return nil, err
	}
	c.Add(pkgs)
	return c.PackageByPath[pkgPath], nil
}
