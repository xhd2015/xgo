package types

import "fmt"

// Type represents type or an object
type Type interface {
	typeMark()
	String() string
	Underlying() UnderlyingType
}

type UnderlyingType interface {
	underlyingTypeMark()
	Type
}

type Basic string
type Unknown struct{}

type ImportPath string

// package level variable
type Variable struct {
	PkgPath string
	Name    string
	Type    Type
}

type VariableField struct {
	Variable Variable
	Field    string
}

type NamedType struct {
	PkgPath string
	Name    string
	Type    Type
}

type Ptr struct {
	Elem Type
}

type Func struct {
	Recv    Type
	Params  []Type
	Results []Type
}

type Struct struct {
	Fields []StructField
}

type StructField struct {
	Name string
	Type Type
}

func (c Basic) typeMark()           {}
func (c Basic) underlyingTypeMark() {}
func (c Basic) Underlying() UnderlyingType {
	return c
}
func (c Basic) String() string {
	return string(c)
}

func (c Unknown) typeMark()           {}
func (c Unknown) underlyingTypeMark() {}
func (c Unknown) String() string {
	return "unknown"
}
func (c Unknown) Underlying() UnderlyingType {
	return c
}

func (c ImportPath) typeMark()           {}
func (c ImportPath) underlyingTypeMark() {}
func (c ImportPath) Underlying() UnderlyingType {
	return c
}
func (c ImportPath) String() string {
	return string(c)
}

func (c Ptr) typeMark()           {}
func (c Ptr) underlyingTypeMark() {}
func (c Ptr) String() string {
	return fmt.Sprintf("*%s", c.Elem.String())
}
func (c Ptr) Underlying() UnderlyingType {
	return c.Elem.Underlying()
}

func (c Variable) typeMark() {}
func (c Variable) Underlying() UnderlyingType {
	return c.Type.Underlying()
}
func (c Variable) String() string {
	return fmt.Sprintf("%s.%s", c.PkgPath, c.Name)
}

func (c VariableField) typeMark()           {}
func (c VariableField) underlyingTypeMark() {}
func (c VariableField) Underlying() UnderlyingType {
	return c.Variable.Underlying()
}
func (c VariableField) String() string {
	return fmt.Sprintf("%s.%s", c.Variable.String(), c.Field)
}

func (c NamedType) typeMark() {}
func (c NamedType) Underlying() UnderlyingType {
	return c.Type.Underlying()
}
func (c NamedType) String() string {
	return fmt.Sprintf("%s.%s", c.PkgPath, c.Name)
}

func IsUnknown(typ Type) bool {
	return typ == Unknown{}
}

func (c Func) typeMark()           {}
func (c Func) underlyingTypeMark() {}
func (c Func) Underlying() UnderlyingType {
	return c
}
func (c Func) String() string {
	return fmt.Sprintf("func(%s) %s", c.Params, c.Results)
}

func (c Struct) typeMark()           {}
func (c Struct) underlyingTypeMark() {}
func (c Struct) Underlying() UnderlyingType {
	return c
}
func (c Struct) String() string {
	return "struct{...}"
}
