package third

type GreetService struct {
}

func (s *GreetService) Greet(name string) string {
	return "hello " + name
}

func Greet(name string) string {
	return "hello " + name
}
