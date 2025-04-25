package constants

const (
	// __xgo_register
	REGISTER_SIGNATURE = "func(v interface{}){}"
	// __xgo_trap
	TRAP_FUNC_SIGNATURE = "func(info interface{}, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool){return nil, false;}"
	// __xgo_trap_var
	TRAP_VAR_SIGNATURE = "func(info interface{}, varAddr interface{}, res interface{}){}"
	// __xgo_trap_varptr
	TRAP_VAR_PTR_SIGNATURE = "func(info interface{}, varAddr interface{}, res interface{}){}"
)
