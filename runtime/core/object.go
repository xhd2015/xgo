package core

type Object interface {
	GetField(name string) Field
	GetFieldIndex(i int) Field
	NumField() int
}

type Field interface {
	Name() string
	Value() interface{}
	Set(val interface{})
}
