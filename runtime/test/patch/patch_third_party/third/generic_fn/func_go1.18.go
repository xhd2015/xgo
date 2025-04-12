//go:build go1.18
// +build go1.18

package generic_fn

func Greet[T ~string](name T) string {
	return "hello " + string(name)
}

func GreetMulti[H ~string, W ~string](h H, w W) string {
	return string(h) + " " + string(w)
}
