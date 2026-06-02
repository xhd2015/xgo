package core

type Object interface {
	GetField(name string) Field
	GetFieldIndex(i int) Field
	NumField() int
}

type ObjectWithErr interface {
	Object

	GetErr() Field
}

type Field interface {
	Name() string
	Value() interface{}
	Ptr() interface{}
	Set(val interface{})
}
