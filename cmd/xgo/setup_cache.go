package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xhd2015/xgo/support/osinfo"
)

const setupManifest = ".xgo-complete.json"

type setupSentinel struct {
	Path string
	Dir  bool
}

type setupRecord struct {
	Revision    string
	GoVersion   string
	FilePatches bool
	Sentinels   []setupSentinel
}

func instrumentSetupRecord(revision, goVersion string, filePatches bool) setupRecord {
	return setupRecord{
		Revision: revision, GoVersion: goVersion, FilePatches: filePatches,
		Sentinels: []setupSentinel{
			{Path: filepath.Join(goVersion, "bin", "go"+osinfo.EXE_SUFFIX)},
			{Path: filepath.Join(goVersion, "pkg", "tool", runtime.GOOS+"_"+runtime.GOARCH, "compile"+osinfo.EXE_SUFFIX)},
			{Path: filepath.Join(goVersion, "src", "runtime", "xgo_trap.go")},
			{Path: filepath.Join(goVersion, "src", "html", "template"), Dir: true},
			{Path: filepath.Join(goVersion, "src", "cmd", "go"), Dir: true},
		},
	}
}

func instrumentRevisionKey(revision string, filePatches bool) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:file-patches=%t:%s/%s", revision, filePatches, runtime.GOOS, runtime.GOARCH))))[:16]
}

// The marker is written last after the private build. Warm validation is
// bounded by a fixed sentinel set, independent of the size of GOROOT. Stat
// follows symlinks, including directory symlinks in a source Go installation.
// Damage outside the sentinels requires an explicit setup --reset; we do not
// retry arbitrary Go failures or rescan the tree on the command's hot path.
func completeInstrumentDir(dir string, expected setupRecord) bool {
	data, err := os.ReadFile(filepath.Join(dir, setupManifest))
	if err != nil {
		return false
	}
	want, err := json.Marshal(expected)
	if err != nil || !bytes.Equal(data, want) {
		return false
	}
	for _, sentinel := range expected.Sentinels {
		info, err := os.Stat(filepath.Join(dir, sentinel.Path))
		if err != nil || info.IsDir() != sentinel.Dir {
			return false
		}
	}
	return len(expected.Sentinels) != 0
}

func writeSetupManifest(dir string, record setupRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, setupManifest), data, 0644)
}

// Return a shared lease held until the child Go command exits. Repair/reset
// waits for all users, builds privately, and publishes only a complete tree.
func prepareInstrumentDir(dir string, record setupRecord, reset bool, build func(string) error) (*cacheLock, error) {
	lease, err := lockCache(dir, false, true)
	if err != nil {
		return nil, err
	}
	if !reset && completeInstrumentDir(dir, record) {
		return lease, nil
	}
	lease.close()
	lease, err = lockCache(dir, true, true)
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			lease.close()
		}
	}()
	if reset || !completeInstrumentDir(dir, record) {
		stage, err := os.MkdirTemp(filepath.Dir(dir), ".xgo-setup-*")
		if err != nil {
			return nil, err
		}
		// Protect staging independently: even a setup taking over an hour
		// must not be mistaken for an abandoned tree by another revision.
		stageLease, err := lockCache(stage, true, true)
		if err != nil {
			os.RemoveAll(stage)
			return nil, err
		}
		defer func() {
			os.RemoveAll(stage)
			stageLease.close()
			// Unlike cache names, random staging names are never reused.
			os.Remove(cacheLockPath(stage))
		}()
		if err := build(stage); err != nil {
			return nil, err
		}
		if err := writeSetupManifest(stage, record); err != nil {
			return nil, err
		}
		if !completeInstrumentDir(stage, record) {
			return nil, fmt.Errorf("incomplete staged toolchain: %s", stage)
		}
		if err := os.RemoveAll(dir); err != nil {
			return nil, err
		}
		if err := os.Rename(stage, dir); err != nil {
			return nil, err
		}
	}
	// Reacquire and validate under a shared lease. flock conversion is not
	// atomic on every Unix; another reset/GC may win the intervening gap.
	lease.close()
	ok = true
	return prepareInstrumentDir(dir, record, false, build)
}

// Source toolchains can be root-owned/read-only. Keep their locks in the
// shared xgo home and canonicalize aliases for both readers and writers.
func sourceGorootLeaseDir(xgoHome, goroot string) (string, error) {
	canonical, err := filepath.EvalSymlinks(goroot)
	if err != nil {
		return "", err
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(canonical))
	return filepath.Join(xgoHome, "goroot-locks", fmt.Sprintf("%x", sum)), nil
}
