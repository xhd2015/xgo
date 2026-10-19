package compiler_extra

type FileDecls struct {
	TrapFuncs      []FuncInfo
	TrapVars       []VarInfo
	InterfaceTypes []InterfaceType
}

type FuncInfo struct {
	IdentityName     string
	Name             string
	HasGenericParams bool
	LineNum          int
	InfoVar          string
	// TrapOnVar is the uint32 flag of a function guarded by __xgo_trap_on_,
	// empty if the function is not guarded
	TrapOnVar string

	RecvPtr      bool
	RecvGeneric  bool
	RecvTypeName string

	Receiver *Field
	Params   Fields
	Results  Fields
}

type VarInfo struct {
	Name    string
	LineNum int
	InfoVar string
}

type InterfaceType struct {
	Name    string
	LineNum int
	InfoVar string
}

type Fields []*Field
type Field struct {
	Name string
}

func (c *Fields) Names() []string {
	names := make([]string, len(*c))
	for i, field := range *c {
		names[i] = field.Name
	}
	return names
}
