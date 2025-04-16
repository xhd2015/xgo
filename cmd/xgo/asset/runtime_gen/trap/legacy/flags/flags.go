package flags

// to inject these flags, see patch/syntax/syntax.go, search for flag_MAIN_MODULE
// this package exists to make flags passed to xgo to be persisted in building and running.

// flag: none
// env: XGO_MAIN_MODULE
// description:
//
//	auto detected by xgo in the beginning of building.
//	can be used prior to ask runtime/debug's info:
//	var mainModulePath = runtime/debug.ReadBuildInfo().Main.Path
const MAIN_MODULE = ""

// flag: --strace
// env: XGO_STACK_TRACE
// description:
//
//	persist the --strace flag when invoking xgo test,
//	if the value is on or true, trace will be automatically
//	collected when test starts and ends
//
// values:
//
//	on,true => --strace, --strace=on, --strace=true
//	off,false,empty string => --strace=off, --strace=false
const STRACE = ""

// flag: --strace-dir
// env: XGO_STACK_TRACE_DIR
// values:
//
//	directory, default current dir
const STRACE_DIR = ""

// flag: --strace-snapshot-main-module
// env: XGO_STRACE_SNAPSHOT_MAIN_MODULE_DEFAULT
// description:
//
//	collecting main module's trace in snapshot mode,
//	while other's are non snapshot mode
//
//	snapshot mode: args are serialized before executing
//	the function, and results are serilaized after return.
//	this is useful if an object will be modified in later
//	process.
//
// values: true or false
const STRACE_SNAPSHOT_MAIN_MODULE_DEFAULT = ""

// flag: --trap-stdlib
// env: XGO_STD_LIB_TRAP_DEFAULT_ALLOW
// description: if true, stdlib trap is by default allowed
// values:
//
//	true - default with test
//	empty string and any other value - --trap-stdlib=false
const TRAP_STDLIB = ""
