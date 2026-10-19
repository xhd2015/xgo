package main

import (
	"os"
	"syscall"
	"unsafe"
)

var lockFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("LockFileEx")
var unlockFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("UnlockFileEx")

func lockFile(f *os.File, exclusive, wait bool) error {
	var flags uintptr
	if exclusive {
		flags |= 2 // LOCKFILE_EXCLUSIVE_LOCK
	}
	if !wait {
		flags |= 1 // LOCKFILE_FAIL_IMMEDIATELY
	}
	var overlapped syscall.Overlapped
	r, _, err := lockFileEx.Call(f.Fd(), flags, 0, 1, 0, uintptr(unsafe.Pointer(&overlapped)))
	if r == 0 {
		return err
	}
	return nil
}

func unlockFile(f *os.File) {
	var overlapped syscall.Overlapped
	unlockFileEx.Call(f.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(&overlapped)))
}
