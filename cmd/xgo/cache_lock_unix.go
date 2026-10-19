//go:build linux || darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"os"
	"syscall"
)

func lockFile(f *os.File, exclusive, wait bool) error {
	op := syscall.LOCK_SH
	if exclusive {
		op = syscall.LOCK_EX
	}
	if !wait {
		op |= syscall.LOCK_NB
	}
	for {
		err := syscall.Flock(int(f.Fd()), op)
		if err != syscall.EINTR {
			return err
		}
	}
}

func unlockFile(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
