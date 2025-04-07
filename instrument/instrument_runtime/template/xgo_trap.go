package template

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

func InstantiateXgoTrap(xgoTrapTemplate string, goVersion *goinfo.GoVersion) (string, error) {
	// instrument xgo_trap.go
	xgoTrapContent, err := patch.RemoveBuildIgnore(xgoTrapTemplate)
	if err != nil {
		return "", fmt.Errorf("remove build ignore: %w", err)
	}
	// type _panic has a major upgrade from go1.21 to go1.22
	// go1.21 and before use pc, go1.22 and after use retpc
	retpc := "retpc"
	if goVersion.Major == 1 && goVersion.Minor <= 21 {
		retpc = "pc"
	}
	xgoTrapContent = strings.ReplaceAll(xgoTrapContent, "__RETPC__", retpc)
	xgoTrapContent = AppendGetFuncNameImpl(goVersion, xgoTrapContent)

	return xgoTrapContent, nil
}
