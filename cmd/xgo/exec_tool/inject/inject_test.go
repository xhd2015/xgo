package inject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInsertRuntimeTrap(t *testing.T) {
	// Create a temporary file with test code
	tempDir, err := os.MkdirTemp("", "inject_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "test.go")
	testCode := `package test

import "fmt"

// A comment before function
func HelloWorld() {
	fmt.Println("Hello, World!")
}

// EmptyFunc is empty
func EmptyFunc() {}

// A struct method
type MyStruct struct{}
func (m MyStruct) Method() {
	fmt.Println("Method")
}

// A pointer receiver
func (m *MyStruct) PointerMethod() {
	fmt.Println("Pointer Method")
}

// A function with parameters and return values
func Add(a, b int) int {
	return a + b
}

// Multiple statements
func MultipleStatements() {
	fmt.Println("First")
	fmt.Println("Second")
}
`

	err = os.WriteFile(testFilePath, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run the function under test
	modified, err := InjectRuntimeTrap(testFilePath)
	if err != nil {
		t.Fatalf("insertRuntimeTrap failed: %v", err)
	}

	// Check the results
	modifiedStr := string(modified)

	// Verify all functions have the trap injected
	expectedFuncs := []string{
		"func HelloWorld() {defer runtime.XgoTrap()();",
		"func EmptyFunc() {defer runtime.XgoTrap()();",
		"func (m MyStruct) Method() {defer runtime.XgoTrap()();",
		"func (m *MyStruct) PointerMethod() {defer runtime.XgoTrap()();",
		"func Add(a, b int) int {defer runtime.XgoTrap()();",
		"func MultipleStatements() {defer runtime.XgoTrap()();",
	}

	for _, expected := range expectedFuncs {
		if !strings.Contains(modifiedStr, expected) {
			t.Errorf("Expected modified code to contain '%s', but it did not", expected)
		}
	}

	// Verify imports are preserved
	if !strings.Contains(modifiedStr, `import "fmt"`) {
		t.Error("Expected imports to be preserved")
	}

	// Verify comments are preserved
	if !strings.Contains(modifiedStr, "// A comment before function") {
		t.Error("Expected comments to be preserved")
	}

	t.Logf("Modified code:\n%s", modifiedStr)
}
