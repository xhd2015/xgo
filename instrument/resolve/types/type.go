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

type Operation interface {
	Info
	operationMark()
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

type OperationName string

type Basic string
type Unknown struct{}
type Untyped struct{}

type ImportPath string

// package level variable
type PkgVariable struct {
	PkgPath string
	Name    string
	Type_   Type
}

// package level function
type PkgFunc struct {
	PkgPath   string
	Name      string
	Signature Signature
}

// package level method
type PkgTypeMethod struct {
	Name      string
	Recv      Type
	Signature Signature
}

type TupleValue []Object

type Tuple []Type

type Pointer struct {
	Value Object
}

type Value struct {
	Type_ Type
}

type Literal struct {
	Type_ Type
}

type ScopeVariable struct {
	Value Object
}

type UntypedNil struct {
}

type NamedType struct {
	PkgPath string
	Name    string
	Type    Type
}

type PtrType struct {
	Elem Type
}

type Lazy func() Info

// LazyType will be resolved when UseContext.doUseTypeInFile is called
// that will recursively resolve the lazy type
type LazyType func() Type

type RawPtr struct {
	Elem UnderlyingType
}

type Signature struct {
	// Params  []Type // don't need it for now
	Results []Type
}

type Struct struct {
	Fields []StructField
}

type StructField struct {
	Name string
	Type Type
}

type Interface struct {
	// Methods []Method
}

type Map struct {
	Key   Type
	Value Type
}

type Slice struct {
	Elem Type
}

type Array struct {
	Elem Type
}

type Chan struct {
	Elem Type
}

func (c OperationName) infoMark()      {}
func (c OperationName) operationMark() {}
func (c OperationName) String() string {
	return string(c)
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

func (c Unknown) infoMark()           {}
func (c Unknown) objectMark()         {}
func (c Unknown) typeMark()           {}
func (c Unknown) underlyingTypeMark() {}
func (c Unknown) Type() Type {
	return c
}
func (c Unknown) Underlying() UnderlyingType {
	return c
}
func (c Unknown) String() string {
	return "unknown"
}

func (c Untyped) infoMark()           {}
func (c Untyped) objectMark()         {}
func (c Untyped) typeMark()           {}
func (c Untyped) underlyingTypeMark() {}
func (c Untyped) Underlying() UnderlyingType {
	return c
}
func (c Untyped) String() string {
	return "untyped"
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

func (c PtrType) infoMark() {}
func (c PtrType) typeMark() {}
func (c PtrType) Underlying() UnderlyingType {
	return RawPtr{Elem: c.Elem.Underlying()}
}
func (c PtrType) String() string {
	return fmt.Sprintf("*%s", c.Elem.String())
}

func (c Lazy) infoMark() {}
func (c Lazy) String() string {
	return "lazy"
}

func (c LazyType) infoMark() {}
func (c LazyType) typeMark() {}
func (c LazyType) Underlying() UnderlyingType {
	return c().Underlying()
}
func (c LazyType) String() string {
	return "lazy_type"
}

func (c RawPtr) infoMark()           {}
func (c RawPtr) typeMark()           {}
func (c RawPtr) underlyingTypeMark() {}
func (c RawPtr) Underlying() UnderlyingType {
	return c
}
func (c RawPtr) String() string {
	return fmt.Sprintf("*%s", c.Elem.String())
}

func (c PkgVariable) infoMark()   {}
func (c PkgVariable) objectMark() {}
func (c PkgVariable) Type() Type {
	return c.Type_
}
func (c PkgVariable) String() string {
	return fmt.Sprintf("%s.%s", c.PkgPath, c.Name)
}

func (c TupleValue) infoMark()   {}
func (c TupleValue) objectMark() {}
func (c TupleValue) Type() Type {
	t := make(Tuple, len(c))
	for i, v := range c {
		t[i] = v.Type()
	}
	return t
}
func (c TupleValue) String() string {
	return "tuple"
}

func (c Pointer) infoMark()   {}
func (c Pointer) objectMark() {}
func (c Pointer) Type() Type {
	return PtrType{Elem: c.Value.Type()}
}
func (c Pointer) String() string {
	return fmt.Sprintf("*%s", c.Value.String())
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

func (c ScopeVariable) infoMark()   {}
func (c ScopeVariable) objectMark() {}
func (c ScopeVariable) Type() Type {
	return c.Value.Type()
}
func (c ScopeVariable) String() string {
	return c.Value.String()
}

func (c UntypedNil) infoMark()   {}
func (c UntypedNil) objectMark() {}
func (c UntypedNil) Type() Type {
	return Untyped{}
}
func (c UntypedNil) String() string {
	return "nil"
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

func (c PkgFunc) infoMark()   {}
func (c PkgFunc) objectMark() {}
func (c PkgFunc) Type() Type {
	return c.Signature
}
func (c PkgFunc) String() string {
	return "func(...)"
}

func (c PkgTypeMethod) infoMark()   {}
func (c PkgTypeMethod) objectMark() {}
func (c PkgTypeMethod) Type() Type {
	return c.Signature
}
func (c PkgTypeMethod) String() string {
	return "func (some)(...)"
}

func (c Signature) infoMark()           {}
func (c Signature) typeMark()           {}
func (c Signature) underlyingTypeMark() {}
func (c Signature) Underlying() UnderlyingType {
	return c
}
func (c Signature) String() string {
	return "func(...)"
}

func (c Tuple) infoMark()           {}
func (c Tuple) typeMark()           {}
func (c Tuple) underlyingTypeMark() {}
func (c Tuple) Underlying() UnderlyingType {
	return c
}
func (c Tuple) Type() Type {
	return c
}
func (c Tuple) String() string {
	return "tuple"
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

func (c Interface) infoMark()           {}
func (c Interface) typeMark()           {}
func (c Interface) underlyingTypeMark() {}
func (c Interface) Underlying() UnderlyingType {
	return c
}
func (c Interface) String() string {
	return "interface{}"
}

func (c Map) infoMark()           {}
func (c Map) typeMark()           {}
func (c Map) underlyingTypeMark() {}
func (c Map) Underlying() UnderlyingType {
	return c
}
func (c Map) String() string {
	return "map[...]"
}

func (c Slice) infoMark()           {}
func (c Slice) typeMark()           {}
func (c Slice) underlyingTypeMark() {}
func (c Slice) Underlying() UnderlyingType {
	return c
}
func (c Slice) String() string {
	return "slice[...]"
}

func (c Array) infoMark()           {}
func (c Array) typeMark()           {}
func (c Array) underlyingTypeMark() {}
func (c Array) Underlying() UnderlyingType {
	return c
}
func (c Array) String() string {
	return fmt.Sprintf("array[...]%s", c.Elem.String())
}

func (c Chan) infoMark()           {}
func (c Chan) typeMark()           {}
func (c Chan) underlyingTypeMark() {}
func (c Chan) Underlying() UnderlyingType {
	return c
}
func (c Chan) String() string {
	return fmt.Sprintf("chan %s", c.Elem.String())
}

func IsPointer(typ Type) bool {
	underlying := typ.Underlying()
	_, ok := underlying.(RawPtr)
	return ok
}

func ResolveLazy(typ Type) Type {
	lzy, ok := typ.(LazyType)
	if !ok {
		return typ
	}
	return lzy()
}

func ResolveAlias(typ NamedType) NamedType {
	for {
		t, ok := typ.Type.(NamedType)
		if !ok {
			return typ
		}
		typ = t
	}
}
