package type_alias

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

type Service struct {
	hello string
}

func (s *Service) Greet(name string) string {
	return fmt.Sprintf("%v %s", s.hello, name)
}

type ServiceString = Service

var svc = ServiceString{hello: "hello"}

func TestTypeAliasNonPtr(t *testing.T) {
	res := svc.Greet("world")
	if res != "hello world" {
		t.Errorf("expect svc.Greet() = %v, but got %v", "hello world", res)
	}
	mock.Patch(&svc, func() *ServiceString {
		return &ServiceString{hello: "mock"}
	})
	result := svc.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect svc.Greet() = %v, but got %v", expected, result)
	}
}
