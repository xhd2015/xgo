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

// when: xgo test,xgo run, xgo build
// flag: --xgo-race-safe
// description:
//
//	skip some functionalities that are not race-safe
//
// values:
//
//	true, on, 1 => --xgo-race-safe, --xgo-race-safe=on, --xgo-race-safe=true, --xgo-race-safe=1
//	false, off, 0, empty string => --xgo-race-safe=off, --xgo-race-safe=false, --xgo-race-safe=0
//
// see https://github.com/xhd2015/xgo/issues/341
const XGO_RACE_SAFE = false
