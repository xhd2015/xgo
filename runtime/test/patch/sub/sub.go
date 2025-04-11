package sub

type NameSet map[string]bool

type Service struct {
	Hello string
}

var NewVar = new(Service)

var PtrVar = &Service{
	Hello: "ptr",
}

var FnService = NewService()

func (c *Service) Greet(s string) string {
	return c.Hello + " " + s
}

func NewService() *Service {
	return &Service{
		Hello: "fn",
	}
}
