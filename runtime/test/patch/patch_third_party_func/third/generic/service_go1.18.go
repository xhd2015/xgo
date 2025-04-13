//go:build go1.18
// +build go1.18

package generic

type GreetService[T string] struct {
}

func (s *GreetService[T]) Greet(name T) string {
	return "hello " + string(name)
}

type GreetMultiService[H string, W string] struct {
}

func (s *GreetMultiService[H, W]) GreetMulti(h H, w W) string {
	return string(h) + " " + string(w)
}
