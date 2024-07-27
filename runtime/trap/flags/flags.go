package flags

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

// flag: --trap-stdlib
// env: XGO_STD_LIB_TRAP_DEFAULT_ALLOW
// description: if true, stdlib trap is by default allowed
// values:
//
//	true - default with test
//	empty string and any other value - --trap-stdlib=false
const TRAP_STDLIB = ""
