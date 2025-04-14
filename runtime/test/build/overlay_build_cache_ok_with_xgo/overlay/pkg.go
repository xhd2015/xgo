package pkg;import "runtime";

func Greet() string { 
	return "hello"+runtime.Version()
}
