//go:build go1.20
// +build go1.20

package func_names

type GI[T any] interface {
	F(T)
}

type GS[T any] struct {
}

func (c *GS[T]) F(T) {

}

func GF[T any](t T) {
}

func init() {
	gs := GS[int]{}

	var gi GI[int] = &gs
	getGenericTests = func() []*testCase {
		return []*testCase{
			{GF[int], "github.com/xhd2015/xgo/test/xgo_test/func_names.GF[...]"},
			{(*GS[int]).F, "github.com/xhd2015/xgo/test/xgo_test/func_names.(*GS[...]).F"},
			{gs.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.(*GS[...]).F-fm"},
			{GI[int].F, "github.com/xhd2015/xgo/test/xgo_test/func_names.GI[...].F"},
			{gi.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.GI[...].F-fm"},
		}
	}
}
