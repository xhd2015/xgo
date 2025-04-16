package instrument_compiler

// gen: go run ./script/generate runtime-def
const RuntimeExtraDef = `
// xgo
func XgoTrap(info interface{}, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool)
func XgoTrapVar(info interface{}, varAddr interface{}, res interface{})
func XgoTrapVarPtr(info interface{}, varAddr interface{}, res interface{})
func XgoRegister(fn interface{})
`
