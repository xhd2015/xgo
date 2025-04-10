//go:build go1.18
// +build go1.18

package generic_fn

func Greet[T ~string](name T) string {
	return "hello " + string(name)
}
