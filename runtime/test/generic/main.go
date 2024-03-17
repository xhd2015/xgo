package main

func main() {
	Hello(1)
}

// funcName1: Hello[go.shape.struct {}]
// funcName2: Hello[main.List[int]]
func Hello[T any](v T) interface{} {
	trapD(&v)
	return v
}

func trapD(x interface{}) (ok bool, after func()) {
	return
}
