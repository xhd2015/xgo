package core

import "testing"

func TestCore(t *testing.T) {
	fi := FuncInfo{Kind: Kind_Func, FullName: "test.Func"}
	if fi.Kind != Kind_Func {
		t.Error("expected Kind_Func")
	}
}
