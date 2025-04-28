// the file will be copied to GOROOT/src/cmd/go/internal/test

//go:build ignore

package test

import (
	"cmd/go/internal/load"
	"fmt"
	"strings"
)

const xgoUnifiedPkgName = "testunified"
const xgoUnifiedPackage = "github.com/xhd2015/xgo/runtime/" + xgoUnifiedPkgName

var xgoTmpDir string

var xgoUnified *load.Package

type xgoTestPackage struct {
	LoadPackage *load.Package
	Tests       []*xgoPkgTestInfo
}

type xgoPkgTestInfo struct {
	Name string // name in t.Run
	Ref  string // name inside the package, i.e. _test_pkg.<Ref>
}

// info mapping when test with json
var xgoTestInfoMapping map[string]xgoTestInfo

type xgoTestInfo struct {
	Package string
	Dir     string
	Ref     string
}

var xgoTestPackages []*xgoTestPackage

// only called when test with json
func xgoGetTestName(pkg string, name string) (pkgPath string, testName string) {
	if pkg != xgoUnifiedPackage {
		return "", ""
	}
	baseName := name
	var suffix string
	idx := strings.Index(name, "/")
	if idx >= 0 {
		baseName = name[:idx]
		suffix = name[idx:]
	}
	info, ok := xgoTestInfoMapping[baseName]
	if !ok {
		return "", name
	}
	return info.Package, info.Ref + suffix
}

func xgoGetPkgAlias(i int) string {
	return fmt.Sprintf("_test_unified_%d", i)
}
