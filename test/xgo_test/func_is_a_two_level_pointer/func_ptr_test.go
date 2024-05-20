//go:build this_is_how_bouke_monkey_works
// +build this_is_how_bouke_monkey_works

// see: https://bou.ke/blog/monkey-patching-in-go/

package func_ptr

import (
	"testing"
	"unsafe"
)

type funcptr struct {
	pc *uintptr
}

type instructions []byte
type funcptrToInstructions struct {
	pc *instructions
}

func Greet(s string) string {
	return "hello " + s
}

func mockGreet(mock func(s string) string) {
	fn := Greet
	x := (*funcptr)(unsafe.Pointer(&fn))
	if false {
		y := (*funcptr)(unsafe.Pointer(&mock))
		*x.pc = *y.pc
	}
	instructions := assembleJump(mock)

	dstInstructions := *((*[]byte)(unsafe.Pointer(x.pc)))
	copy(dstInstructions, instructions)
}

func assembleJump(f func(s string) string) []byte {
	funcVal := *(*uintptr)(unsafe.Pointer(&f))
	return []byte{
		0x48, 0xC7, 0xC2,
		byte(funcVal >> 0),
		byte(funcVal >> 8),
		byte(funcVal >> 16),
		byte(funcVal >> 24), // MOV rdx, funcVal
		0xFF, 0x22,          // JMP [rdx]
	}
}

func TestMock(t *testing.T) {
	mockGreet(func(s string) string {
		return "mock " + s
	})

	s := Greet("world")
	t.Logf("s: %v", s)
}
