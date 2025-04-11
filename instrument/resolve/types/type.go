package types

import "fmt"

type Info interface {
	infoMark()
}

// Object represents type or an object
type Object interface {
	Info
	objectMark()
	String() string
	Type() Type
}

type Type interface {
	Info
	String() string
	Underlying() UnderlyingType
	typeMark()
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
	Type_   Type
}

type Value struct {
	Type_ Type
}

type Literal struct {
	Type_ Type
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
	PkgPath string
	Name    string

	Signature Signature
}

type Method struct {
	Name      string
	Recv      Type
	Signature Signature
}

type Signature struct {
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

func (c Basic) infoMark()           {}
func (c Basic) typeMark()           {}
func (c Basic) underlyingTypeMark() {}
func (c Basic) Underlying() UnderlyingType {
	return c
}
func (c Basic) String() string {
	return string(c)
}

func (c Unknown) infoMark()   {}
func (c Unknown) objectMark() {}
func (c Unknown) typeMark()   {}
func (c Unknown) Type() Type {
	panic("should not call Type() on unknown")
}
func (c Unknown) Underlying() UnderlyingType {
	panic("should not call Type() on unknown")
}
func (c Unknown) String() string {
	return "unknown"
}

func (c ImportPath) infoMark()           {}
func (c ImportPath) typeMark()           {}
func (c ImportPath) underlyingTypeMark() {}
func (c ImportPath) Underlying() UnderlyingType {
	return c
}
func (c ImportPath) String() string {
	return string(c)
}

func (c Ptr) infoMark()           {}
func (c Ptr) typeMark()           {}
func (c Ptr) underlyingTypeMark() {}
func (c Ptr) String() string {
	return fmt.Sprintf("*%s", c.Elem.String())
}
func (c Ptr) Underlying() UnderlyingType {
	return c.Elem.Underlying()
}

func (c Variable) infoMark()   {}
func (c Variable) objectMark() {}
func (c Variable) Type() Type {
	return c.Type_
}
func (c Variable) String() string {
	return fmt.Sprintf("%s.%s", c.PkgPath, c.Name)
}

func (c Value) infoMark()   {}
func (c Value) objectMark() {}
func (c Value) Type() Type {
	return c.Type_
}
func (c Value) String() string {
	return fmt.Sprintf("%s{}", c.Type_.String())
}

func (c Literal) infoMark()   {}
func (c Literal) objectMark() {}
func (c Literal) Type() Type {
	return c.Type_
}
func (c Literal) String() string {
	return fmt.Sprintf("%s{}", c.Type_.String())
}

func (c NamedType) infoMark() {}
func (c NamedType) typeMark() {}
func (c NamedType) Underlying() UnderlyingType {
	return c.Type.Underlying()
}
func (c NamedType) String() string {
	return fmt.Sprintf("%s.%s", c.PkgPath, c.Name)
}

func IsUnknown(info Info) bool {
	return info == Unknown{}
}

func (c Func) infoMark()   {}
func (c Func) objectMark() {}
func (c Func) Type() Type {
	return c.Signature
}
func (c Func) String() string {
	return "func(...)"
}

func (c Method) infoMark()   {}
func (c Method) objectMark() {}
func (c Method) Type() Type {
	return c.Signature
}
func (c Method) String() string {
	return "func (some)(...)"
}

func (c Signature) infoMark()           {}
func (c Signature) typeMark()           {}
func (c Signature) underlyingTypeMark() {}
func (c Signature) Underlying() UnderlyingType {
	return c
}
func (c Signature) String() string {
	return fmt.Sprintf("func(...)")
}

func (c Struct) infoMark()           {}
func (c Struct) typeMark()           {}
func (c Struct) underlyingTypeMark() {}
func (c Struct) Underlying() UnderlyingType {
	return c
}
func (c Struct) String() string {
	return "struct{...}"
}
