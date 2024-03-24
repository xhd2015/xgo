package func_list

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/functab"
)

const testPkgPath = "github.com/xhd2015/xgo/runtime/test/func_list"

var addExtraPkgsAssert func(m map[string]bool)

func TestFuncList(t *testing.T) {
	funcs := functab.GetFuncs()

	total := 0
	countByPkg := make(map[string]int)

	missingPkgs := map[string]bool{
		testPkgPath + ".example":       true,
		testPkgPath + ".someInt.value": true,
		testPkgPath + ".someInt.inc":   true,
	}
	if addExtraPkgsAssert != nil {
		addExtraPkgsAssert(missingPkgs)
	}
	for _, fn := range funcs {
		pkg := fn.Pkg
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

	// for pkgName, count := range countByPkg {
	// 	fmt.Printf(" >> %s: %v\n", pkgName, count)
	// }

	// fmt.Printf("total: %d\n", total)
}

func example() {

}

type someInt int

func (c someInt) value() int {
	return int(c)
}
func (c *someInt) inc() int {
	*c++
	return int(*c)
}
