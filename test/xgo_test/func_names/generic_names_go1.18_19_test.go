//go:build go1.18 && !go1.20
// +build go1.18,!go1.20

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
			{GF[int], "github.com/xhd2015/xgo/test/xgo_test/func_names.init.0.func1.1"},
			{(*GS[int]).F, "github.com/xhd2015/xgo/test/xgo_test/func_names.init.0.func1.2"},
			// it seems that with go1.18,go1.19 ,gs.F does not end with -fm suffix
			// that may cause generic mock failed.
			{gs.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.init.0.func1.3"},
			{GI[int].F, "github.com/xhd2015/xgo/test/xgo_test/func_names.GI[...].F"},
			{gi.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.GI[...].F-fm"},
		}
	}
}
