package gen_defs

type GenernateType string

const (
	// auto increment number
	GenernateType_CmdXgoVersion GenernateType = "cmd/xgo/version.go"

	// copy from cmd/xgo/version.go CORE_VERSION* to runtime/core/version.go
	GenernateType_RuntimeCoreVersion GenernateType = "runtime/core/version.go"

	// copy from runtime to cmd/xgo/trace/render/stack_model/stack_model.go
	GenernateType_RuntimeTraceModel GenernateType = "runtime/trace/stack_model/stack_model.go"

	// copy from runtime to cmd/xgo/runtime_gen
	GenernateType_XgoRuntimeGen GenernateType = "cmd/xgo/runtime_gen"
	// copy from cmd/xgo/upgrade
	GenernateType_ScriptInstallUpgrade GenernateType = "script/install/upgrade"

	// copy from runtime/core/func_template.go to runtime/core/func.go
	GenernateType_RuntimeCoreFunc GenernateType = "runtime/core/func.go"

	// copy from runtime/internal/runtime/xgo_trap_template.go to runtime/internal/runtime/xgo_trap_template.go
	GenernateType_RuntimeXgoTrapTemplate GenernateType = "runtime/internal/runtime/xgo_trap_template.go"
)
