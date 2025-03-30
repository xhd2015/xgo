//go:build go1.18
// +build go1.18

package mock_generic

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/trap"
)

// go run ./cmd/xgo test --project-dir runtime -run TestInspectPCGeneric -v ./test/mock_generic
func TestInspectPCGeneric(t *testing.T) {
	_, fnInfo, funcPC, trappingPC := trap.InspectPC(ToString[int])
	_, fnInfoStr, funcPCStr, trappingPCStr := trap.InspectPC(ToString[string])

	if !fnInfo.Generic {
		t.Fatalf("expect ToString[int] to be generic")
	}

	if false {
		// debug
		t.Logf("funcPC: 0x%x", funcPC)
		t.Logf("trappingPC: 0x%x", trappingPC)
		t.Logf("funcPCStr: 0x%x", funcPCStr)
		t.Logf("trappingPCStr: 0x%x", trappingPCStr)
	}

	if fnInfo != fnInfoStr {
		t.Fatalf("expect ToString[int] and ToString[string] returns the same func info")
	}

	if funcPC == trappingPC {
		t.Fatalf("expect ToString[int] pc to be different, actually the same: funcPC=0x%x, trappingPC=0x%x", funcPC, trappingPC)
	}
	if funcPCStr == trappingPCStr {
		t.Fatalf("expect ToString[string] pc to be different, actually the same: funcPC=0x%x, trappingPC=0x%x", funcPCStr, trappingPCStr)
	}
	if funcPC == funcPCStr {
		t.Fatalf("expect ToString[int] and ToString[string] to have different pc, actually the same: funcPC=0x%x, funcPCStr=0x%x", funcPC, funcPCStr)
	}
	if trappingPC == trappingPCStr {
		t.Fatalf("expect ToString[int] and ToString[string] to have different trapping PC, actually the same: trappingPC=0x%x, trappingPCStr=0x%x", trappingPC, trappingPCStr)
	}
}
