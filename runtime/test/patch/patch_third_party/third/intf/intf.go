package intf

type IGreet interface {
	Greet(name string) string
}

type greetService struct {
}

func (s *greetService) Greet(name string) string {
	return "hello " + name
}

func NewGreetService() IGreet {
	return &greetService{}
}
