// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	_ "encoding/json"
	_ "io"
	_ "io/ioutil"
	_ "net"
	_ "net/http"
	_ "os/exec"
	"testing"
	_ "time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

func TestListStdlib(t *testing.T) {
	funcs := functab.GetFuncs()

	stdPkgs := map[string]bool{
		"net/http.Get": true,
	}
	found, missing := getMissing(funcs, stdPkgs, false)
	if len(missing) > 0 {
		t.Fatalf("expect func list contains: %v, actual %v", missing, found)
	}
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
		// debug
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
