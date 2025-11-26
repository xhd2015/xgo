package filter_trace

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

// TestFilterTraceByName tests that FilterTrace can filter functions by name
func TestFilterTraceByName(t *testing.T) {
	var capturedStack stack_model.IStack

	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			capturedStack = stack
		},
		FilterTrace: func(funcInfo *core.FuncInfo) bool {
			// Only trace functions that contain "Include" in their name
			return strings.Contains(funcInfo.Name, "Include")
		},
	}, nil, func() (interface{}, error) {
		// Call functions directly so they're at the same depth level
		functionInclude1()
		functionExclude1()
		return nil, nil
	})

	if capturedStack == nil {
		t.Fatal("expected stack to be captured")
	}

	data, err := capturedStack.JSON()
	if err != nil {
		t.Fatalf("failed to marshal stack: %v", err)
	}

	dataStr := string(data)

	// Should include functionInclude1 (and its children)
	if !strings.Contains(dataStr, "functionInclude1") {
		t.Errorf("expected trace to contain functionInclude1, got: %s", dataStr)
	}

	// Should NOT include functionExclude1 (filtered out)
	if strings.Contains(dataStr, "functionExclude1") {
		t.Errorf("expected trace to NOT contain functionExclude1, got: %s", dataStr)
	}
}

// TestFilterTraceByPackage tests that FilterTrace can filter functions by package
func TestFilterTraceByPackage(t *testing.T) {
	var capturedStack stack_model.IStack

	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			capturedStack = stack
		},
		FilterTrace: func(funcInfo *core.FuncInfo) bool {
			// Only trace functions in this test package
			return strings.Contains(funcInfo.Pkg, "filter_trace")
		},
	}, nil, func() (interface{}, error) {
		functionB()
		return nil, nil
	})

	if capturedStack == nil {
		t.Fatal("expected stack to be captured")
	}

	data, err := capturedStack.JSON()
	if err != nil {
		t.Fatalf("failed to marshal stack: %v", err)
	}

	dataStr := string(data)

	// Should include our test functions
	if !strings.Contains(dataStr, "functionB") {
		t.Errorf("expected trace to contain functionB, got: %s", dataStr)
	}
}

// TestFilterTraceNil tests that when FilterTrace is nil, all functions are traced
func TestFilterTraceNil(t *testing.T) {
	var capturedStack stack_model.IStack

	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			capturedStack = stack
		},
		FilterTrace: nil, // No filter, should trace everything
	}, nil, func() (interface{}, error) {
		functionC()
		return nil, nil
	})

	if capturedStack == nil {
		t.Fatal("expected stack to be captured")
	}

	data, err := capturedStack.JSON()
	if err != nil {
		t.Fatalf("failed to marshal stack: %v", err)
	}

	dataStr := string(data)

	// Should include all functions
	if !strings.Contains(dataStr, "functionC") {
		t.Errorf("expected trace to contain functionC, got: %s", dataStr)
	}
	if !strings.Contains(dataStr, "helperC1") {
		t.Errorf("expected trace to contain helperC1, got: %s", dataStr)
	}
	if !strings.Contains(dataStr, "helperC2") {
		t.Errorf("expected trace to contain helperC2, got: %s", dataStr)
	}
}

// TestFilterTraceExcludeAll tests that FilterTrace can exclude all functions
func TestFilterTraceExcludeAll(t *testing.T) {
	var capturedStack stack_model.IStack

	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			capturedStack = stack
		},
		FilterTrace: func(funcInfo *core.FuncInfo) bool {
			// Exclude all functions
			return false
		},
	}, nil, func() (interface{}, error) {
		functionD()
		return nil, nil
	})

	if capturedStack == nil {
		t.Fatal("expected stack to be captured")
	}

	data, err := capturedStack.JSON()
	if err != nil {
		t.Fatalf("failed to marshal stack: %v", err)
	}

	dataStr := string(data)

	// Should NOT include any of our test functions (only the trace wrapper might be there)
	if strings.Contains(dataStr, "functionD") {
		t.Errorf("expected trace to NOT contain functionD when all filtered out, got: %s", dataStr)
	}
	if strings.Contains(dataStr, "helperD1") {
		t.Errorf("expected trace to NOT contain helperD1 when all filtered out, got: %s", dataStr)
	}
}

// TestFilterTraceComplex tests a more complex filtering scenario
func TestFilterTraceComplex(t *testing.T) {
	var capturedStack stack_model.IStack

	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			capturedStack = stack
		},
		FilterTrace: func(funcInfo *core.FuncInfo) bool {
			// Include functions that:
			// 1. Are not from stdlib
			// 2. Have "E" in their name
			if funcInfo.Stdlib {
				return false
			}
			return strings.Contains(funcInfo.Name, "E")
		},
	}, nil, func() (interface{}, error) {
		// Call functions at the same level
		helperE1()
		helperWithoutMatch()
		return nil, nil
	})

	if capturedStack == nil {
		t.Fatal("expected trace to be captured")
	}

	data, err := capturedStack.JSON()
	if err != nil {
		t.Fatalf("failed to marshal stack: %v", err)
	}

	// Parse to check structure
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	dataStr := string(data)

	// t.Logf("dataStr: %s", dataStr)

	// Should include helperE1 (has "E")
	if !strings.Contains(dataStr, "helperE1") {
		t.Errorf("expected trace to contain helperE1, got: %s", dataStr)
	}

	// Should NOT include helperWithoutMatch (doesn't have "E")
	if strings.Contains(dataStr, "helperWithoutMatch") {
		t.Errorf("expected trace to NOT contain helperWithoutMatch, got: %s", dataStr)
	}
}

// TestFilterTraceDeep tests a more complex filtering scenario
func TestFilterTraceDeep(t *testing.T) {
	var capturedStack stack_model.IStack

	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			capturedStack = stack
		},
		FilterTrace: func(funcInfo *core.FuncInfo) bool {
			if funcInfo.Stdlib || funcInfo.Closure {
				return false
			}
			if funcInfo.IdentityName == "deep2" {
				return false
			}
			return true
		},
	}, nil, func() (interface{}, error) {
		// Call functions at the same level
		deepCall()
		return nil, nil
	})

	if capturedStack == nil {
		t.Fatal("expected trace to be captured")
	}

	data, err := capturedStack.JSON()
	if err != nil {
		t.Fatalf("failed to marshal stack: %v", err)
	}

	if !strings.Contains(string(data), "deepCall") {
		t.Errorf("expected trace to contain deepCall, got: %s", string(data))
	}
	if !strings.Contains(string(data), "deep1") {
		t.Errorf("expected trace to contain deep1, got: %s", string(data))
	}
	if !strings.Contains(string(data), "deep1A") {
		t.Errorf("expected trace to contain deep1A, got: %s", string(data))
	}
	if !strings.Contains(string(data), "deep3") {
		t.Errorf("expected trace to contain deep3, got: %s", string(data))
	}

	// should not contain deep2
	if strings.Contains(string(data), "deep2") {
		t.Errorf("expected trace to NOT contain deep2, got: %s", string(data))
	}

}

// Helper functions for tests

func functionInclude1() string {
	functionInclude2()
	return "include1"
}

func functionInclude2() string {
	return "include2"
}

func functionExclude1() string {
	functionExclude2()
	return "exclude1"
}

func functionExclude2() string {
	return "exclude2"
}

func functionB() {
	helperB1()
	helperB2()
}

func helperB1() string {
	return "b1"
}

func helperB2() string {
	return "b2"
}

func functionC() {
	helperC1()
	helperC2()
}

func helperC1() string {
	return "c1"
}

func helperC2() string {
	return "c2"
}

func functionD() {
	helperD1()
}

func helperD1() string {
	return "d1"
}

func helperE1() string {
	return "e1"
}

func helperWithoutMatch() string {
	return "no match"
}

func deepCall() {
	deep1()
	deep1A()
}

func deep1() {
	deep2()
}

func deep1A() {
	deep1()
}

func deep2() {
	deep3()
}

func deep3() {

}
