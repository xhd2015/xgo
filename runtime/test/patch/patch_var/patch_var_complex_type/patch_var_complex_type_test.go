package patch_var_complex_type

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

type StructType struct {
	ID int
}

type RecursiveType struct {
	ID    int
	Child *RecursiveType
}

var strctVar = StructType{
	ID: 1,
}

var strctVarPtr = &StructType{
	ID: 2,
}

var recursiveVar = RecursiveType{
	ID: 3,
	Child: &RecursiveType{
		Child: nil,
	},
}

func TestPatchStructVarValue(t *testing.T) {
	mock.Patch(&strctVar, func() StructType {
		return StructType{
			ID: 100,
		}
	})
	after := strctVar
	if after.ID != 100 {
		t.Fatalf("expect strctVar.ID to be %v, actual: %v", 100, after.ID)
	}
}

func TestPatchStructVarPtr(t *testing.T) {
	mock.Patch(&strctVarPtr, func() *StructType {
		return &StructType{
			ID: 100,
		}
	})
	after := strctVarPtr
	if after.ID != 100 {
		t.Fatalf("expect strctVarPtr.ID to be %v, actual: %v", 100, after.ID)
	}
}

func TestPatchRecursiveTypeValue(t *testing.T) {
	mock.Patch(&recursiveVar, func() RecursiveType {
		return RecursiveType{
			ID: 100,
		}
	})
	after := recursiveVar
	if after.ID != 100 {
		t.Fatalf("expect recursiveVar.ID to be %v, actual: %v", 100, after.ID)
	}
}

func TestPatchRecursiveTypePtr(t *testing.T) {
	mock.Patch(&recursiveVar, func() *RecursiveType {
		return &RecursiveType{
			ID: 100,
		}
	})
	after := &recursiveVar
	if after.ID != 100 {
		t.Fatalf("expect recursiveVar.ID to be %v, actual: %v", 100, after.ID)
	}
}
