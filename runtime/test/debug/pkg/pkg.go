package pkg

import "fmt"

// format: pkg.Func, pkg.Recv.Method, pkg.(*Recv).Method

func Hello(p string) {
	fmt.Printf("hello %s\n", p)

	m := Mass(1)

	// github.com/xhd2015/xgo/runtime/pkg.Mass.Print
	m.Print("g")

	person := &Person{Name: "test"}
	// github.com/xhd2015/xgo/runtime/pkg.(*Person).Greet
	person.Greet("runtime")

	// github.com/xhd2015/xgo/runtime/pkg.Hello.func1
	func() {
		fmt.Printf("I'm from unnamed closure\n")
	}()
}

type Mass int

func (c Mass) Print(suffix string) {
	fmt.Printf("Mass: %d%s\n", c, suffix)
}

type Person struct {
	Name string
}

func (c *Person) Greet(who string) {
	fmt.Printf("Person: greet %s->%s\n", c.Name, who)
}
