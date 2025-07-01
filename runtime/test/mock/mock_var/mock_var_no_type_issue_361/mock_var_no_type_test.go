package mock_var_no_type

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// go run ./script/run-test -tags=dev --include go1.24.2 --debug-xgo ./runtime/test/mock/mock_var/mock_var_no_type_issue_361

type Cat struct {
	Name string
	Hat  *Hat
}
type Hat struct {
}

func (c *Cat) GetName() string {
	return c.Name
}

var tom = &Cat{Name: "tom"}
var tom2 = Cat{Name: "tom2"}

// name has type string
var name = tom.Name

// hat has type *Hat
var hat = tom.Hat

// name2 has type string
var name2 = tom2.Name

func TestStructPatch(t *testing.T) {
	fmt.Println(tom)
	fmt.Println(name)

	jack := &Cat{
		Name: "jack",
	}
	mock.Patch(jack.GetName, func() string {
		return tom.Name
	})

	if n := jack.GetName(); n != "tom" {
		t.Fatalf("expect GetName() to be 'tom', actual: %s", n)
	}
}
