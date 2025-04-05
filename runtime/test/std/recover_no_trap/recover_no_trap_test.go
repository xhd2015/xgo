package recover_no_trap

import (
	"context"
	"testing"
	"text/template"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

// bug see https://github.com/xhd2015/xgo/issues/164
// flags: --trap-stdlib
func TestRecoverInStdlibShouldCapturePanic(t *testing.T) {
	// if the implementation is right, no panic should happen
	txt := `test {{define "content"}}`

	_, err := template.New("").Parse(txt)
	expectMsg := "template: :1: unexpected EOF"
	if err == nil || err.Error() != expectMsg {
		t.Fatalf("expect parse err: %q, actual: %q", expectMsg, err.Error())
	}
}

func TestRecoverInNonStdlibShouldBeTrapped(t *testing.T) {
	var haveMockedA bool
	mock.Mock(A, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		haveMockedA = true
		return nil
	})
	A()
	if !haveMockedA {
		t.Fatalf("expect have mocked A, actually not")
	}

	var mockBSetupErr interface{}
	var haveMockedB bool
	var result string
	func() {
		defer func() {
			mockBSetupErr = recover()
		}()
		mock.Mock(B, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
			haveMockedB = true
			return nil
		})
		result = B()
	}()

	if mockBSetupErr != nil {
		t.Fatalf("expect setup mock B no error, actual: %v", mockBSetupErr)
	}
	if !haveMockedB {
		t.Fatalf("expect haveMockedB to be true, actual: false")
	}
	if result != "" {
		t.Fatalf("expect B() returns mocked empty string, actual: %v", result)
	}
}
