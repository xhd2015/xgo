package sub

var A string = "subA"

type Mapping map[int]string

func (c Mapping) Get(i int) string {
	return c[i]
}
