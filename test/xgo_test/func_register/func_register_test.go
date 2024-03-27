package func_register

import (
	"reflect"
	"testing"
)

func init() {
	__xgo_link_on_init_finished(initFuncMapping)
}

func __xgo_link_retrieve_all_funcs_and_clear(f func(fn interface{})) {
	panic("should be linked by compiler")
}

func __xgo_link_on_init_finished(f func()) {
	panic("should be linked by compiler")
}

type fnInfo struct {
	PkgPath      string
	IdentityName string
}

var fnInfos []*fnInfo

func initFuncMapping() {
	__xgo_link_retrieve_all_funcs_and_clear(func(fn interface{}) {
		rv := reflect.ValueOf(fn)
		pkgPath := rv.FieldByName("PkgPath").String()
		idName := rv.FieldByName("IdentityName").String()
		fnInfos = append(fnInfos, &fnInfo{
			PkgPath:      pkgPath,
			IdentityName: idName,
		})
	})
}

// TODO: add test on closure register
// go run ./cmd/xgo test -v -run TestFuncRegister ./test/xgo_test/func_register
func TestFuncRegister(t *testing.T) {
	var hasF bool
	for _, fn := range fnInfos {
		// debug
		t.Logf("fn: %v %v", fn.PkgPath, fn.IdentityName)

		if fn.PkgPath == "github.com/xhd2015/xgo/test/xgo_test/func_register" && fn.IdentityName == "F" {
			hasF = true
		}
	}
	if !hasF {
		t.Fatalf("expect has F, actual not found")
	}
}

func F() {}
