package main

import (
	"os"
	"path/filepath"
)

// Locks live beside, never inside, the directory they protect. Removing or
// publishing a cache must not replace the inode on which other runs wait.
// Lock files are intentionally retained, including after GC.
type cacheLock struct {
	file *os.File
}

func lockCache(dir string, exclusive, wait bool) (*cacheLock, error) {
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(cacheLockPath(dir), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f, exclusive, wait); err != nil {
		f.Close()
		return nil, err
	}
	return &cacheLock{file: f}, nil
}

func (l *cacheLock) close() {
	if l != nil && l.file != nil {
		unlockFile(l.file)
		l.file.Close()
		l.file = nil
	}
}

func cacheLockPath(dir string) string {
	return filepath.Join(filepath.Dir(dir), "."+filepath.Base(dir)+".lock")
}
