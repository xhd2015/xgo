package generic

type GreetService[T string] struct {
}

func (s *GreetService[T]) Greet(name T) string {
	return "hello " + string(name)
}
