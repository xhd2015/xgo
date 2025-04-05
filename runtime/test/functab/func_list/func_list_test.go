package func_list

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

const testPkgPath = "github.com/xhd2015/xgo/runtime/test/functab/func_list"

var addExtraPkgsAssert func(m map[string]bool)

// go run ./script/run-test/ --include go1.17.13 --xgo-runtime-test-only -run TestFuncListFn -v ./test/func_list
// go run ./cmd/xgo test --project-dir runtime -run TestFuncListFn -v ./test/func_list
func TestFuncListFn(t *testing.T) {
	funcs := functab.GetFuncs()

	missingPkgs := map[string]bool{
		testPkgPath + ".example":     true,
		testPkgPath + ".someStr.get": true,
		testPkgPath + ".someStr.set": true,
		// interface will not be included
		// testPkgPath + ".someIntf.GetName":           true,
		// testPkgPath + ".someEmbedIntf.GetName":      true,
		// testPkgPath + ".someEmbedIntf.GetEmbedName": true,
	}
	missingIntfs := map[string]bool{
		testPkgPath + ".someIntf":      true,
		testPkgPath + ".someEmbedIntf": true,
	}
	if addExtraPkgsAssert != nil {
		addExtraPkgsAssert(missingPkgs)
	}

	found, missing := getMissing(funcs, missingPkgs, false)
	if len(missing) > 0 {
		t.Fatalf("expect func list contains: %v, actual %v", missing, found)
	}

	foundIntf, missIntf := getMissing(funcs, missingIntfs, true)
	if len(missIntf) > 0 {
		t.Fatalf("expect func list contains interfaces: %v, actual %v", missIntf, foundIntf)
	}
	// for pkgName, count := range countByPkg {
	// 	fmt.Printf(" >> %s: %v\n", pkgName, count)
	// }

	// fmt.Printf("total: %d\n", total)
}

func getMissing(funcs []*core.FuncInfo, missingPkgs map[string]bool, intf bool) (found []string, missing []string) {
	for _, fn := range funcs {
		pkg := fn.Pkg
		if intf {
			if fn.Interface {
				// t.Logf("found interface: %v", fn)
				key := pkg + "." + fn.RecvType
				if _, ok := missingPkgs[key]; ok {
					missingPkgs[key] = false
				}
			}
			continue
		}
		displayName := fn.DisplayName()
		// fmt.Printf("found: %s %s\n", pkg, displayName)
		if pkg == testPkgPath {
			// debug
			// t.Logf("found: %s %s", pkg, displayName)
		}
		key := pkg + "." + displayName
		_, ok := missingPkgs[key]
		if ok {
			missingPkgs[key] = false
		}
	}
	for k, ok := range missingPkgs {
		if ok {
			missing = append(missing, k)
		} else {
			found = append(found, k)
		}
	}
	return
}

func example() {

}

type someStr string

func (c someStr) get() string {
	return string(c)
}
func (c *someStr) set(val string) string {
	*c = someStr(val)
	return val
}

type someIntf interface {
	GetName() string
}

type someEmbedIntf interface {
	someIntf
	GetEmbedName() string
}
