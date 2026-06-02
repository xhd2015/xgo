//go:build go1.22
// +build go1.22

package pkg_with_source_files_with_5_files_with_4_decls

// The overlay injects 4 synthetic init blocks (__xgo_register_1,
// __xgo_trap_1, __xgo_trap_var_1, __xgo_trap_varptr_1), consuming
// closure slots 1-4. The user's top-level closure var topClosure
// therefore lands at slot 5.
const expectTopLevelClosure = "github.com/xhd2015/xgo/runtime/test/func_names/pkg_with_source_files_with_5_files_with_4_decls.init.func5"
