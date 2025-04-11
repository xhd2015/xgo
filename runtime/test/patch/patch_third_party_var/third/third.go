package third

type GreetService struct {
	Hello string
}

func (s *GreetService) Greet(name string) string {
	return s.Hello + " " + name
}

func NewService() *GreetService {
	return &GreetService{
		Hello: "hello",
	}
}
