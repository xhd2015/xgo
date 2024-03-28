// reflect is important to implement mock.MockMethodByName

package reflect_type_info

import (
	"reflect"
	"testing"

	"github.com/xhd2015/xgo/test/xgo_test/reflect_type_info/sub"
)

// go run ./cmd/xgo test -run TestReflectTypePkgPathDifferentOnExportedness -v ./test/xgo_test/reflect_type_info
// Description:
// - exported type: have pkg path
// - unexported type: have no pkg path
func TestReflectTypePkgPathDifferentOnExportedness(t *testing.T) {
	s1 := &struct_{}
	sv := reflect.ValueOf(s1)

	expectPkgPath := ""
	svType := sv.Type()
	if svType.PkgPath() != expectPkgPath {
		t.Fatalf("expect svType.PkgPath() to be %s, actual: %s", expectPkgPath, svType.PkgPath())
	}

	expectExpPkg := "github.com/xhd2015/xgo/test/xgo_test/reflect_type_info"
	exportedPkgPath := reflect.TypeOf(Struct_{}).PkgPath()
	if exportedPkgPath != expectExpPkg {
		t.Fatalf("expect TypeOf(Struct_{}).PkgPath() to be %s, actual: %s", expectExpPkg, exportedPkgPath)
	}
}

func TestGetUnexportedMethodOfStruct(t *testing.T) {
	s1 := &struct_{}

	sv := reflect.ValueOf(s1)

	_ = sv
}

// go test -run TestGetUnexportedMethodOfInterface -v ./test/xgo_test/reflect_type_info
func TestGetUnexportedMethodOfInterface(t *testing.T) {
	lowi := sub.GetLowInterface_()

	lowv := reflect.ValueOf(lowi)

	// lowv is a ptr?
	lowiKind := lowv.Kind()
	t.Logf("kind: %s", lowiKind.String())

	lowM1 := lowv.MethodByName("M1")
	if !lowM1.IsValid() {
		t.Fatalf("expect to get lowv.M1, actual nil")
	}
}

type Struct_ struct {
}

type struct_ struct {
	field func()
}

func (c *struct_) m1() {

}

func (c struct_) m2() {

}

func (c struct_) M3() {

}

type interface_ interface {
	m1()
	m2()
	M3()
}
