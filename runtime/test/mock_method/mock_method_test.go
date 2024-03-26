package mock_method

import (
	"context"
	"fmt"
	"testing"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

type struct_ struct {
	name  string
	value int
}

func (c *struct_) String() string {
	return fmt.Sprintf("<%s>: %v", c.name, c.value)
}

// go run ./cmd/xgo test --project-dir runtime -run TestMethodMockOnPtr -v ./test/mock_method
func TestMethodMockOnPtr(t *testing.T) {
	s1 := &struct_{
		name:  "s1",
		value: 1,
	}
	s2 := &struct_{
		name:  "s2",
		value: 2,
	}
	s1StrExpect := "<s1>: 1"
	s1Str := s1.String()
	if s1Str != s1StrExpect {
		t.Fatalf("expect s1Str to be %s, actual: %s", s1StrExpect, s1Str)
	}
	s2StrExpect := "<s2>: 2"
	s2Str := s2.String()
	if s2Str != s2StrExpect {
		t.Fatalf("expect s2Str to be %s, actual: %s", s2StrExpect, s2Str)
	}

	mockStr := "mock str"
	cancel := mock.Mock(s1.String, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set(mockStr)
		return nil
	})

	s1StrAfterMock := s1.String()
	if s1StrAfterMock != mockStr {
		t.Fatalf("expect s1StrAfterMock to be %s, actual: %s", mockStr, s1StrAfterMock)
	}

	s2StrAfterMock := s2.String()
	if s2StrAfterMock != s2StrExpect {
		t.Fatalf("expect s1StrAfterMock to be %s, actual: %s", s2StrExpect, s2StrAfterMock)
	}

	cancel()
	s1StrAfterMockCancel := s1.String()
	if s1StrAfterMockCancel != s1Str {
		t.Fatalf("expect s1StrAfterMock to be %s, actual: %s", s1Str, s1StrAfterMockCancel)
	}
}

// no recver name
func (*struct_) Type() string {
	return "struct_"
}

// empty name
func (_ struct_) TypePtr() string {
	return "*struct_"
}

// go run ./cmd/xgo test --project-dir runtime -run TestMethodMockBlankName -v ./test/mock_method
func TestMethodMockBlankName(t *testing.T) {
	s1 := &struct_{
		name:  "s1",
		value: 1,
	}
	s1Size := unsafe.Sizeof(*s1)
	nameSize := unsafe.Sizeof(s1.name)
	valueSize := unsafe.Sizeof(s1.value)
	if s1Size != nameSize+valueSize {
		t.Fatalf("expect sizeof(struct_) to be %d, actual: %d", nameSize+valueSize, s1Size)
	}
	s1TypExpect := "struct_"
	s1Typ := s1.Type()
	if s1Typ != s1TypExpect {
		t.Fatalf("expect s1Typ to be %s, actual: %s", s1TypExpect, s1Typ)
	}

	s1TypPtrExpect := "*struct_"
	s1TypPtr := s1.TypePtr()
	if s1TypPtr != s1TypPtrExpect {
		t.Fatalf("expect s1TypPtr to be %s, actual: %s", s1TypPtrExpect, s1TypPtr)
	}

	mockType := "mock type"
	cancel := mock.Mock(s1.Type, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set(mockType)
		return nil
	})

	s1TypAfterMock := s1.Type()
	if s1TypAfterMock != mockType {
		t.Fatalf("expect s1TypAfterMock to be %s, actual: %s", mockType, s1TypAfterMock)
	}

	cancel()

	s1TypAfterMockCancel := s1.Type()
	if s1TypAfterMockCancel != s1TypExpect {
		t.Fatalf("expect s1TypAfterMockCancel to be %s, actual: %s", s1TypExpect, s1TypAfterMockCancel)
	}
}

type interface_ interface {
	String() string
}

// go run ./script/run-test/ --include go1.17.13 --xgo-runtime-test-only -run TestMethodMockOnInterface -v ./test/mock_method
func TestMethodMockOnInterface(t *testing.T) {
	s1 := &struct_{
		name:  "s1",
		value: 1,
	}
	var intf interface_ = s1

	mockStr := "mock interface"
	mock.Mock(intf.String, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set(mockStr)
		return nil
	})
	res := intf.String()
	if res != mockStr {
		t.Fatalf("expect mock interface_.String() to be %s, actual: %s", mockStr, res)
	}
}
