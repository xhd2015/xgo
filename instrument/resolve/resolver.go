package resolve

import "github.com/xhd2015/xgo/instrument/edit"

// Resolver resolves type of an AST Expr.
// a type is defined as giving the following information:
// - package path
// - type name
// - has a receivier
// - is the receiver a pointer
// - is the receiver generic
//
// the result type are split into 3 parts:
// - package-level variable
// - package-level function
// - package-level method of a type
type Resolver struct {
}

func NewResolver() *Resolver {
	return &Resolver{}
}

// a -> look for definition
// a.b -> look for a's definition first, a could a package name, a type or a variable.
// a.b.c -> look for a.b's definition
func (r *Resolver) Resolve(packages *edit.Packages) error {
	return nil
}

// TODO: can we support generic functions and methods even in go1.18 and go1.19 with the
// help of resolver?
// through a rewritting of mock.Patch? If we know the function has been instrumented,
// we record that to some place, and replace the call with mock.AutoGenPatchGeneric(which is generated on the fly), in that function the pc lookup is done by inspecting directly.
// or more generally, we generate the
