package func_list

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/functab"
)

const testPkgPath = "github.com/xhd2015/xgo/runtime/test/func_list"

var addExtraPkgsAssert func(m map[string]bool)

// go run ./script/run-test/ --include go1.17.13 --xgo-runtime-test-only -run TestFuncList -v func_list
func TestFuncList(t *testing.T) {
	funcs := functab.GetFuncs()

	total := 0
	countByPkg := make(map[string]int)

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
	for _, fn := range funcs {
		pkg := fn.Pkg
		if fn.Interface {
			// t.Logf("found interface: %v", fn)
			key := pkg + "." + fn.RecvType
			if _, ok := missingIntfs[key]; ok {
				missingIntfs[key] = false
			}
		}
		displayName := fn.DisplayName()
		if pkg == testPkgPath {
			// debug
			// t.Logf("found: %s %s", pkg, displayName)
		}
		key := pkg + "." + displayName
		_, ok := missingPkgs[key]
		if ok {
			missingPkgs[key] = false
		}
		total++
		countByPkg[pkg]++
	}
	var missing []string
	var found []string
	for k, ok := range missingPkgs {
		if ok {
			missing = append(missing, k)
		} else {
			found = append(found, k)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("expect func list contains: %v, actual %v", missing, found)
	}

	var missIntf []string
	var foundIntf []string
	for k, ok := range missingIntfs {
		if ok {
			missIntf = append(missIntf, k)
		} else {
			foundIntf = append(foundIntf, k)
		}
	}

	if len(missIntf) > 0 {
		t.Fatalf("expect func list contains interfaces: %v, actual %v", missIntf, foundIntf)
	}
	// for pkgName, count := range countByPkg {
	// 	fmt.Printf(" >> %s: %v\n", pkgName, count)
	// }

	// fmt.Printf("total: %d\n", total)
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
