//go:build go1.25
// +build go1.25

package func_names

// go1.25 sets XGO_COMPILER_SYNTAX_REWRITE_ENABLE=true, which triggers
// the patched compiler's AfterFilesParsed to generate __xgo_autogen with
// 4 package-level closures (__xgo_register_0, __xgo_trap_0, __xgo_trap_var_0,
// __xgo_trap_varptr_0). These consume init.func1-4, pushing the user's
// top-level closure var c = func(){} to init.func5.
const expectTopLevelFunc = "github.com/xhd2015/xgo/test/xgo_test/func_names.init.func5"
