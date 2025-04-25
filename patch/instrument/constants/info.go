package constants

import (
	"strconv"
)

const (
	LINK_REGISTER     = "__xgo_register_"
	LINK_TRAP_FUNC    = "__xgo_trap_"
	LINK_TRAP_VAR     = "__xgo_trap_var_"
	LINK_TRAP_VAR_PTR = "__xgo_trap_varptr_"
)

const (
	FUNC_INFO_TYPE = "__xgo_func_info_"
	PKG_VAR        = "__xgo_pkg_"
	FILE_VAR       = "__xgo_file_"
	// for variable
	FILE_VAR_FOR_VAR = "__xgo_file_var_"
	// for compiler extra
	FILE_VAR_GC = "__xgo_file_gc_"
	INIT_FUNC   = "__xgo_init_"
)

const (
	FUNC_INFO = "__xgo_func_info" // __xgo_func_info_<fileIndex>_<declIndex>
	VAR_INFO  = "__xgo_var_info"  // __xgo_var_info_<fileIndex>_<declIndex>
	INTF_INFO = "__xgo_intf_info" // __xgo_intf_info_<fileIndex>_<declIndex>
)

const (
	REG_FILE_GEN = "__xgo_reg_file_gen_"
)

type InfoKind int

const (
	InfoKind_Func   InfoKind = 0
	InfoKind_Var    InfoKind = 1
	InfoKind_VarPtr InfoKind = 2
	InfoKind_Const  InfoKind = 3
)

var FUNC_INFO_FIELDS = []string{
	"Kind int",
	"FullName string",
	"Pkg string",
	"IdentityName string",
	"Name string",
	"RecvType string",
	"RecvPtr bool",
	"Interface bool",
	"Generic bool",
	"Closure bool",
	"Stdlib bool",
	"File string",
	"Line int",
	"PC uintptr",
	"Func interface{}",
	"Var interface{}",
	"RecvName string",
	"ArgNames []string",
	"ResNames []string",
	"FirstArgCtx bool",
	"LastResultErr bool",
}

func Register(fileIndex int) string {
	return LINK_REGISTER + strconv.Itoa(fileIndex)
}

func Trap(fileIndex int) string {
	return LINK_TRAP_FUNC + strconv.Itoa(fileIndex)
}

func TrapVar(fileIndex int) string {
	return LINK_TRAP_VAR + strconv.Itoa(fileIndex)
}

func TrapVarPtr(fileIndex int) string {
	return LINK_TRAP_VAR_PTR + strconv.Itoa(fileIndex)
}

func FuncInfoType(fileIndex int) string {
	return FUNC_INFO_TYPE + strconv.Itoa(fileIndex)
}

func PkgVar(fileIndex int) string {
	return PKG_VAR + strconv.Itoa(fileIndex)
}

func FileVar(fileIndex int) string {
	return FILE_VAR + strconv.Itoa(fileIndex)
}

func FileVarForVar(fileIndex int) string {
	return FILE_VAR_FOR_VAR + strconv.Itoa(fileIndex)
}

func FileVarGc(fileIndex int) string {
	return FILE_VAR_GC + strconv.Itoa(fileIndex)
}

func InitFunc(fileIndex int) string {
	return INIT_FUNC + strconv.Itoa(fileIndex)
}

func FuncInfoVarName(fileIndex int, declIndex int) string {
	return FUNC_INFO + "_" + strconv.Itoa(fileIndex) + "_" + strconv.Itoa(declIndex)
}

func VarInfoVarName(fileIndex int, declIndex int) string {
	return VAR_INFO + "_" + strconv.Itoa(fileIndex) + "_" + strconv.Itoa(declIndex)
}

func IntfInfoVarName(fileIndex int, declIndex int) string {
	return INTF_INFO + "_" + strconv.Itoa(fileIndex) + "_" + strconv.Itoa(declIndex)
}

func RegFileGen(fileIndex int) string {
	return REG_FILE_GEN + strconv.Itoa(fileIndex)
}
