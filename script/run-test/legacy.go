package main

// TODO: remove duplicate test between xgo test and runtime test
var runtimeSubTests = []string{
	"func_list",
	"trap",
	"trap_inspect_func",
	"trap_args",
	"mock_func",
	"mock_method",
	"mock_by_name",
	"mock_closure",
	"mock_stdlib",
	"mock_generic",
	"mock_var",
	"patch",
	"patch_const",
	"tls",
}

type TestCase struct {
	name         string
	usePlainGo   bool //  use go instead of xgo
	dir          string
	flags        []string
	windowsFlags []string
	env          []string
	skipOnCover  bool

	skipOnTimeout     bool
	windowsFailIgnore bool
}

// can be selected via --name
// use:
//
//	go run ./scrip/run-test --list
//
// to list all names
var extraSubTests = []*TestCase{
	{
		name:  "trace_without_dep",
		dir:   "runtime/test/trace_without_dep",
		flags: []string{"--strace"},
		// see https://github.com/xhd2015/xgo/issues/144#issuecomment-2138565532
		windowsFlags: []string{"--trap-stdlib=false", "--strace"},
	},
	{
		name:         "trace_without_dep_vendor",
		dir:          "runtime/test/trace_without_dep_vendor",
		flags:        []string{"--strace"},
		windowsFlags: []string{"--trap-stdlib=false", "--strace"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/87
		name:         "trace_without_dep_vendor_replace",
		dir:          "runtime/test/trace_without_dep_vendor_replace",
		flags:        []string{"--strace"},
		windowsFlags: []string{"--trap-stdlib=false", "--strace"},
	},
	// trap
	{
		name: "trap_flags_persistent",
		dir:  "runtime/test/trap/flags/persistent_after_build",
	},
	{
		name:  "trap_with_overlay",
		dir:   "runtime/test/trap_with_overlay",
		flags: []string{"-overlay", "overlay.json"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/111
		name:  "trap_stdlib_any",
		dir:   "runtime/test/trap_stdlib_any",
		flags: []string{"--trap-stdlib"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/142
		name:  "bugs_regression",
		dir:   "runtime/test/bugs/...",
		flags: []string{},
	},
	{
		name:         "trace_marshal",
		dir:          "runtime/test/trace_marshal/...",
		flags:        []string{},
		windowsFlags: []string{"--trap-stdlib=false"},
	},
	{
		name:         "trace",
		dir:          "runtime/test/trace",
		flags:        []string{},
		windowsFlags: []string{"--trap-stdlib=false"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/142
		name:         "trace",
		dir:          "runtime/test/trace/check_trace_flag",
		flags:        []string{},
		windowsFlags: []string{"--trap-stdlib=false"},
	},
	{
		// go run ./script/run-test/ --name trace-marshal-not-trap-stdlib
		name:  "trace-marshal-not-trap-stdlib",
		dir:   "runtime/test/trace/marshal/flag",
		flags: []string{"--trap-stdlib=false"},
	},
	{
		name:  "trace-marshal-exclude",
		dir:   "runtime/test/trace/marshal/exclude",
		flags: []string{"--mock-rule", `{"pkg":"encoding/json","name":"newTypeEncoder","action":"exclude"}`},
	},
	{
		name:          "trace-snapshot",
		dir:           "runtime/test/trace/snapshot",
		skipOnCover:   true,
		skipOnTimeout: true,
	},
	{
		name:              "trace-custom-dir",
		dir:               "runtime/test/trace/trace_dir",
		windowsFailIgnore: true,
	},
	{
		// see https://github.com/xhd2015/xgo/issues/202
		name: "asm_func",
		dir:  "runtime/test/issue_194_asm_func",
	},
	{
		// see https://github.com/xhd2015/xgo/issues/194
		name: "asm_func_sonic",
		dir:  "runtime/test/issue_194_asm_func/demo",
	},
	{
		// see https://github.com/xhd2015/xgo/issues/142
		name:         "trace_panic_peek",
		dir:          "runtime/test/trace_panic_peek/...",
		flags:        []string{},
		windowsFlags: []string{"--trap-stdlib=false"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/164
		name:  "stdlib_recover_no_trap",
		dir:   "runtime/test/recover_no_trap",
		flags: []string{"--trap-stdlib"},
	},
	{
		name:  "mock_rule_not_set",
		dir:   "runtime/test/mock/rule",
		flags: []string{"-run", "TestClosureDefaultMock"},
	},
	{
		name:  "mock_rule_set",
		dir:   "runtime/test/mock/rule",
		flags: []string{"-run", "TestClosureWithMockRuleNoMock", "--mock-rule", `{"closure":true,"action":"exclude"}`},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/231
		name:  "cross_build",
		dir:   "runtime/test/build/simple",
		flags: []string{"-c", "-o", "/dev/null"},
		env:   []string{"GOOS=", "GOARCH="},
	},
	{
		name:       "xgo_integration",
		usePlainGo: true,
		dir:        "test/xgo_integration",
	},
	{
		name:       "timeout",
		usePlainGo: true,
		dir:        "runtime/test/timeout",
		// the test is 600ms sleep
		flags:             []string{"-timeout=0.2s"},
		skipOnTimeout:     true,
		windowsFailIgnore: true,
	},
}
