package flags

// when: xgo test
// flag: --strace
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
const COLLECT_TEST_TRACE = false

// when: xgo test and --strace is on
// flag: --strace-dir
// values:
//
//	directory, default current dir
const COLLECT_TEST_TRACE_DIR = ""
