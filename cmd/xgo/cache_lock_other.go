//go:build !linux && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !windows

package main

import (
	"fmt"
	"os"
	"time"
)

// On platforms without native advisory locks, serialize readers too. Never
// reclaim an apparently stale lock: a paused process may still be using it.
// After an unclean exit, the .held file must be removed manually.
func lockFile(f *os.File, exclusive, wait bool) error {
	deadline := time.Now().Add(time.Minute)
	for {
		h, err := os.OpenFile(f.Name()+".held", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			return h.Close()
		}
		if !os.IsExist(err) || !wait {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for cache lock %s.held; if no xgo process is using it, remove that file manually", f.Name())
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func unlockFile(f *os.File) {
	os.Remove(f.Name() + ".held")
}
