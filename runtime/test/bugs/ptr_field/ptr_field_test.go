package ptr_field

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

type DummyConfig struct {
	DummyMap map[string]string
}

var dummyConfigPtr *DummyConfig
var dummyConfig DummyConfig

func InitConfigPtr() {
	// expected to be &dummyConfigPtr_xgo_get().DummyMap
	err := json.Unmarshal([]byte(`{"key": "original"}`), &dummyConfigPtr.DummyMap)
	if err != nil {
		panic(err)
	}
}

func InitConfig() {
	// expected to be skipped?
	err := json.Unmarshal([]byte(`{"key": "original"}`), &dummyConfig.DummyMap)
	if err != nil {
		panic(err)
	}
}

func TestSetupMockForDummyConfigPtrShouldWork(t *testing.T) {
	// This test demonstrates that mocking dummyConfigPtr (a pointer variable)
	// works correctly even when it's used in &dummyConfigPtr.DummyMap pattern.
	// The variable is trapped and rewritten to use _xgo_get() which returns *DummyConfig.

	// Initialize the pointer variable
	dummyConfigPtr = &DummyConfig{}

	InitConfigPtr()

	// Verify the original value before mocking
	originalValue := dummyConfigPtr.DummyMap["key"]
	if originalValue != "original" {
		t.Fatalf("expect original value to be 'original', got: %s", originalValue)
	}

	mockCalled := false
	mock.Mock(&dummyConfigPtr, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		mockCalled = true
		results.GetFieldIndex(0).Set(&DummyConfig{
			DummyMap: map[string]string{
				"key": "mocked",
			},
		})
		return nil
	})

	// Access the variable to trigger the mock
	result := dummyConfigPtr

	// The mock should be called because dummyConfigPtr is trapped
	if !mockCalled {
		t.Fatalf("mock should be called for pointer variable")
	}

	// Verify the mocked value
	if result.DummyMap["key"] != "mocked" {
		t.Fatalf("expect mocked value to be 'mocked', got: %s", result.DummyMap["key"])
	}
}

func TestSetupMockForDummyConfigAlsoWorks(t *testing.T) {
	// This test demonstrates that mocking dummyConfig (a non-pointer variable)
	// also works when it's used directly.

	InitConfig()

	// Verify the original value before mocking
	originalValue := dummyConfig.DummyMap["key"]
	if originalValue != "original" {
		t.Fatalf("expect original value to be 'original', got: %s", originalValue)
	}

	mockCalled := false
	mock.Mock(&dummyConfig, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		mockCalled = true
		// Return a mocked config with different value
		results.GetFieldIndex(0).Set(DummyConfig{
			DummyMap: map[string]string{
				"key": "mocked",
			},
		})
		return nil
	})

	// Access the variable to trigger the mock
	result := dummyConfig

	// The mock should be called because dummyConfig is trapped when used directly
	if !mockCalled {
		t.Fatalf("mock should be called for variable used directly")
	}

	// Verify the mocked value
	if result.DummyMap["key"] != "mocked" {
		t.Fatalf("expect mocked value to be 'mocked', got: %s", result.DummyMap["key"])
	}
}
