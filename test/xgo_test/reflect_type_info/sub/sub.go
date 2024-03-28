package sub

type Interface_ interface {
	M()
	m()
}

type interface_ interface {
	M1()
	m1()
}

type struct_ struct {
}

func (c *struct_) M1() {}
func (c *struct_) m1() {}

func GetLowInterface_() interface_ {
	return &struct_{}
}
