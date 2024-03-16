package main

func main() {
	var list List[int]
	list.Size()

	Hello(list)
}

type List[T any] struct {
}

func (c *List[T]) Size() {

}

// funcName1: Hello[go.shape.struct {}]
// funcName2: Hello[main.List[int]]
func Hello[T any](v T) interface{} {
	return v
}
