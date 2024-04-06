//go:build !windows
// +build !windows

package osinfo

const EXE_SUFFIX = ""

// when copy files, should use
// symbolic as long as possible
const FORCE_COPY_UNSYM = false

const IS_WINDOWS = true
