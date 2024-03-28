package mock_closuer

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/mock_by_name/sub"
)

const subPkg = "github.com/xhd2015/xgo/runtime/test/mock_by_name/sub"

// go run ./cmd/xgo test --project-dir runtime -run TestMockByName -v ./test/mock_by_name
// go run ./script/run-test/ --include go1.22.1 --xgo-runtime-test-only -run TestMockByName -v ./test/mock_by_name
func TestMockByNameExported(t *testing.T) {
	shouldPanic(sub.F)

	mock.MockByName(subPkg, "F", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})

	sub.F()
}

func TestMockByNameUnexported(t *testing.T) {
	shouldPanic(sub.Call_f)

	mock.MockByName(subPkg, "f", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})

	sub.Call_f()
}

func TestMockByNameStructExported(t *testing.T) {
	shouldPanic(sub.Call_sF)

	mock.MockByName(subPkg, "struct_.F", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})

	sub.Call_sF()
}

func TestMockByNameStructUnexported(t *testing.T) {
	shouldPanic(sub.Call_sf)

	mock.MockByName(subPkg, "struct_.f", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})

	sub.Call_sf()
}

func TestMockInstanceMethodByNameUnexported(t *testing.T) {
	shouldPanic(sub.Call_sf)

	s := sub.GetS()
	mock.MockMethodByName(s, "f", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})
	// Call_sf_instance()
	sub.Call_sf()
}

// the struct is empty, so all instances are the same
// go run ./cmd/xgo test --project-dir runtime -run TestMockInstanceMethodByNameUnexportedShouldNotAffectOtherInstance -v ./test/mock_by_name
func TestMockInstanceMethodByNameUnexportedShouldAffectOtherEmptyInstances(t *testing.T) {
	shouldPanic(sub.Call_sf)

	s := sub.GetS()
	s2 := sub.GetS()
	if s != s2 {
		t.Fatalf("expect empty struct to be the same")
	}
	mock.MockMethodByName(s, "f", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})
	sub.Call_sf()
	sub.Call_sf_instance(s)
}

// the struct is not empty
// go run ./cmd/xgo test --project-dir runtime -run TestMockInstanceMethodByNameUnexportedShouldNotAffectOtherInstance -v ./test/mock_by_name
func TestMockInstanceMethodByNameUnexportedShouldNotAffectOtherInstance(t *testing.T) {
	shouldPanic(sub.Call_sf)

	s1 := sub.GetNS("s1")
	s2 := sub.GetNS("s2")
	if s1 == s2 {
		t.Fatalf("expect non-empty struct to be different")
	}
	mock.MockMethodByName(s1, "f", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})
	// s1 mocked, so no panic
	sub.Call_nsf_instance(s1)

	// s2 not mocked, so panic
	shouldPanic(func() {
		sub.Call_nsf_instance(s2)
	})
}

func shouldPanic(f func()) {
	defer func() {
		e := recover()
		if e == nil {
			panic("expect panic, actually not")
		}
	}()
	f()
}
