package instrument_reg

const (
	REGISTER_SIGNATURE     = "func(v interface{}){}"
	TRAP_FUNC_SIGNATURE    = "func(info interface{}, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool){return nil, false;}"
	TRAP_VAR_SIGNATURE     = "func(info interface{}, varAddr interface{}, res interface{}){}"
	TRAP_VAR_PTR_SIGNATURE = "func(info interface{}, varAddr interface{}, res interface{}){}"
)
