package patch

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestPatchUnmatchSignature(t *testing.T) {
	s1 := capturePanic(func() {
		mock.Patch((*struct_).greet, func() string {
			return ""
		})
	})
	expectS1 := "replacer should have type: func(*patch.struct_) string, actual: func() string"
	if s1 != expectS1 {
		t.Fatalf("expect s1 to be %q, actual: %q", expectS1, s1)
	}

	s2 := capturePanic(func() {
		s := &struct_{}
		mock.Patch(s.greet, func(*struct_) string {
			return ""
		})
	})
	expectS2 := "replacer should have type: func() string, actual: func(*patch.struct_) string"
	if s2 != expectS2 {
		t.Fatalf("expect s2 to be %q, actual: %q", expectS2, s2)
	}
}

func TestPatchMethodByNameUnmatchSignature(t *testing.T) {
	s1 := capturePanic(func() {
		s := &struct_{}
		mock.PatchMethodByName(s, "greet", func(*struct_) string {
			return ""
		})
	})
	expectS1 := "replacer should have type: func() string, actual: func(*patch.struct_) string"
	if s1 != expectS1 {
		t.Fatalf("expect s1 to be %q, actual: %q", expectS1, s1)
	}

	s2 := capturePanic(func() {
		mock.PatchByName("github.com/xhd2015/xgo/runtime/test/patch", "(*struct_).greet", func() string {
			return ""
		})
	})
	expectS2 := "replacer should have type: func(*patch.struct_) string, actual: func() string"
	if s2 != expectS2 {
		t.Fatalf("expect s2 to be %q, actual: %q", expectS2, s2)
	}
}

func capturePanic(fn func()) (s string) {
	defer func() {
		e := recover()
		if e != nil {
			s = fmt.Sprint(e)
		}
	}()
	fn()
	return
}
