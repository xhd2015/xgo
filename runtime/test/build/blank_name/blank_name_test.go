package blank_name

import (
	"testing"
)

func TestBlankName(t *testing.T) {
	var c int
	_ = c
	// build pass
}

// these two functions were skipped during noder

func _() {

}

func _(int) {

}
func _(_ int) {

}

// with go1.20: arg[0].Nname=nil
func a(int) {

}

// with go1.20: arg[0].Nname has value, but arg[0].Sym.Name="_"
func b(_ int) {

}

func c() int {
	return 0
}

func d() (_ int) {
	return
}

func e() (i int) {
	return
}

type A struct {
}

type _ struct {
}

func (*A) _() {

}

func (_ *A) a() {

}

func (*A) b() {

}

func (*A) c() {

}

func (*A) d(int) {

}

func (*A) e(_ int, s string) {

}

func (*A) f(x int, s string) {

}
func (*A) g(a int) int {
	return 0
}

func (*A) h(a int) (b int) {
	return
}
