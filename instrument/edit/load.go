package edit

import (
	"github.com/xhd2015/xgo/instrument/load"
)

func (c *Packages) Load(pkgPath string) (*Package, bool, error) {
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
