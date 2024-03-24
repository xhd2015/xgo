//go:build windows
// +build windows

package osinfo

const EXE_SUFFIX = ".exe"

// when copy files, don't use
// symbolic as it may cause failure
const FORCE_COPY_UNSYM = false
