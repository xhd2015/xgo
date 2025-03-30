package sub

func F() {
	panic("F should be mocked")
}

func f() {
	panic("f should be mocked")
}

type struct_ struct {
}

func (c *struct_) F() {
	panic("*struct_.F should be mocked")
}

func (c *struct_) f() {
	panic("*struct_.f should be mocked")
}

type nstruct_ struct {
	name string
}

func (c *nstruct_) F() {
	panic("*nstruct_.F should be mocked")
}

func (c *nstruct_) f() {
	panic("*nstruct_.f should be mocked")
}

func Call_f() {
	f()
}

func Call_sF() {
	e := &struct_{}
	e.F()
}

func Call_sf() {
	e := &struct_{}
	e.f()
}

func GetS() interface{} {
	return &struct_{}
}

func GetNS(name string) interface{} {
	return &nstruct_{
		name: name,
	}
}

func Call_sf_instance(s interface{}) {
	s.(*struct_).f()
}

func Call_nsf_instance(s interface{}) {
	s.(*nstruct_).f()
}
