package patch_var_struct_field_ptr

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestGetProvider(t *testing.T) {
	var called bool
	mock.Patch(&pm, func() *ProviderManager {
		called = true
		return &ProviderManager{}
	})
	res := getProvider("test")
	if !called {
		t.Fatal("should call newProvider")
	}
	if res != "test" {
		t.Fatal("should return test")
	}
}

func TestGetFactoryProvider(t *testing.T) {
	var called bool
	mock.Patch(&pf, func() *ProviderFactory {
		called = true
		return &ProviderFactory{}
	})
	res := getFactoryprovider("test")
	if !called {
		t.Fatal("should call newProvider")
	}
	if res != nil {
		t.Fatal("should return nil")
	}
}
