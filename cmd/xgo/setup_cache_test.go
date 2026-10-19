package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheLockProcess(t *testing.T) {
	if dir := os.Getenv("XGO_LOCK_TEST_DIR"); dir != "" {
		lease, err := lockCache(dir, true, false)
		if err == nil {
			lease.close()
			os.Exit(2)
		}
		return
	}
	dir := filepath.Join(t.TempDir(), "cache")
	lease, err := lockCache(dir, false, true)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.close()
	cmd := exec.Command(os.Args[0], "-test.run=^TestCacheLockProcess$")
	cmd.Env = append(os.Environ(), "XGO_LOCK_TEST_DIR="+dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child did not observe lease: %v: %s", err, out)
	}
	lease.close()
	exclusive, err := lockCache(dir, true, false)
	if err != nil {
		t.Fatal(err)
	}
	exclusive.close()
}

func TestPrepareInstrumentDirConcurrentAndRepair(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "goroot")
	var builds int32
	build := func(stage string) error {
		atomic.AddInt32(&builds, 1)
		if err := os.MkdirAll(filepath.Join(stage, "src", "html", "template"), 0755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(stage, "src", "html", "template", "template.go"), []byte("package template"), 0644)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lease, err := prepareInstrumentDir(dir, testSetupRecord("revision", filepath.Join("src", "html", "template", "template.go")), false, build)
			if err != nil {
				t.Error(err)
				return
			}
			defer lease.close()
			if !completeInstrumentDir(dir, testSetupRecord("revision", filepath.Join("src", "html", "template", "template.go"))) {
				t.Error("published incomplete tree")
			}
		}()
	}
	wg.Wait()
	if builds != 1 {
		t.Fatalf("built %d times", builds)
	}
	if err := os.RemoveAll(filepath.Join(dir, "src", "html")); err != nil {
		t.Fatal(err)
	}
	lease, err := prepareInstrumentDir(dir, testSetupRecord("revision", filepath.Join("src", "html", "template", "template.go")), false, build)
	if err != nil {
		t.Fatal(err)
	}
	lease.close()
	if builds != 2 || !completeInstrumentDir(dir, testSetupRecord("revision", filepath.Join("src", "html", "template", "template.go"))) {
		t.Fatal("did not repair partially deleted tree")
	}
	if completeInstrumentDir(dir, testSetupRecord("other-revision", filepath.Join("src", "html", "template", "template.go"))) {
		t.Fatal("accepted wrong revision")
	}
}

func TestPrepareInstrumentDirFailureAndReset(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "goroot")
	build := func(stage string) error {
		return os.WriteFile(filepath.Join(stage, "file"), []byte("ready"), 0644)
	}
	lease, err := prepareInstrumentDir(dir, testSetupRecord("r", "file"), false, build)
	if err != nil {
		t.Fatal(err)
	}
	lease.close()
	_, err = prepareInstrumentDir(dir, testSetupRecord("r", "file"), true, func(stage string) error {
		return fmt.Errorf("interrupted setup")
	})
	if err == nil || !completeInstrumentDir(dir, testSetupRecord("r", "file")) {
		t.Fatal("failed staging build damaged published tree")
	}
	entries, err := os.ReadDir(filepath.Dir(dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() != "goroot" {
			t.Fatalf("leaked stage: %s", e.Name())
		}
	}
}

func TestGCSkipsLiveLease(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "go_current")
	live := filepath.Join(parent, "go_live")
	for _, dir := range []string{current, live} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	markUsed(live, now.Add(-2*staleBuildCacheAge))
	lease, err := lockCache(live, false, true)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.close()
	units := []gcUnit{{dir: current, namePrefix: "go", maxAge: staleBuildCacheAge}}
	gcStaleCaches(units, now)
	if _, err := os.Stat(live); err != nil {
		t.Fatalf("GC deleted live directory: %v", err)
	}
	lease.close()
	gcStaleCaches(units, now.Add(2*gcInterval))
	if _, err := os.Stat(live); !os.IsNotExist(err) {
		t.Fatalf("GC did not collect released directory: %v", err)
	}
}

func TestInstrumentRevisionKey(t *testing.T) {
	if instrumentRevisionKey("a", true) == instrumentRevisionKey("b", true) ||
		instrumentRevisionKey("a", true) == instrumentRevisionKey("a", false) {
		t.Fatal("setup variants must coexist")
	}
}

func TestResetWaitsForConsumer(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "goroot")
	build := func(stage string) error {
		return os.WriteFile(filepath.Join(stage, "file"), []byte("ready"), 0644)
	}
	lease, err := prepareInstrumentDir(dir, testSetupRecord("r", "file"), false, build)
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		next, err := prepareInstrumentDir(dir, testSetupRecord("r", "file"), true, func(stage string) error {
			close(entered)
			return build(stage)
		})
		if next != nil {
			next.close()
		}
		done <- err
	}()
	select {
	case <-entered:
		t.Error("reset started while a consumer held a lease")
	case <-time.After(50 * time.Millisecond):
	}
	lease.close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("reset did not resume after consumer exited")
	}
}

func testSetupRecord(revision, path string) setupRecord {
	return setupRecord{Revision: revision, Sentinels: []setupSentinel{{Path: path}}}
}

func TestSetupSentinels(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "target"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "target", "file"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink("target", link); err != nil {
		t.Skip(err)
	}
	record := setupRecord{Revision: "r", GoVersion: "go1.27.2", FilePatches: true, Sentinels: []setupSentinel{{Path: "link", Dir: true}, {Path: filepath.Join("link", "file")}}}
	if err := writeSetupManifest(dir, record); err != nil {
		t.Fatal(err)
	}
	if !completeInstrumentDir(dir, record) {
		t.Fatal("directory symlink rejected")
	}
	changed := record
	changed.FilePatches = false
	if completeInstrumentDir(dir, changed) {
		t.Fatal("accepted incompatible patch mode")
	}
	changed = record
	changed.GoVersion = "go1.28.0"
	if completeInstrumentDir(dir, changed) {
		t.Fatal("accepted incompatible Go version")
	}
	if err := os.Remove(filepath.Join(dir, "target", "file")); err != nil {
		t.Fatal(err)
	}
	if completeInstrumentDir(dir, record) {
		t.Fatal("accepted missing sentinel")
	}
}

func TestGCAbandonedSetup(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "go_current")
	now := time.Now()
	for _, name := range []string{"go_current", ".xgo-setup-old", ".xgo-setup-live", ".xgo-setup-new"} {
		dir := filepath.Join(parent, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		when := now.Add(-2 * staleSetupAge)
		if name == ".xgo-setup-new" {
			when = now
		}
		if err := os.Chtimes(dir, when, when); err != nil {
			t.Fatal(err)
		}
	}
	live := filepath.Join(parent, ".xgo-setup-live")
	lease, err := lockCache(live, true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.close()
	units := []gcUnit{{dir: current, namePrefix: "go", maxAge: staleInstrumentDirAge}}
	gcStaleCaches(units, now)
	if _, err := os.Stat(filepath.Join(parent, ".xgo-setup-old")); !os.IsNotExist(err) {
		t.Fatal("abandoned stage retained", err)
	}
	for _, name := range []string{".xgo-setup-live", ".xgo-setup-new"} {
		if _, err := os.Stat(filepath.Join(parent, name)); err != nil {
			t.Fatal("collected active or young stage", err)
		}
	}
	lease.close()
	gcStaleCaches(units, now.Add(2*gcInterval))
	if _, err := os.Stat(live); !os.IsNotExist(err) {
		t.Fatal("released abandoned stage retained", err)
	}
}

func TestSetupMarkerIsBounded(t *testing.T) {
	dir := t.TempDir()
	record := instrumentSetupRecord("revision", "go1.27.2", true)
	for _, sentinel := range record.Sentinels {
		path := filepath.Join(dir, sentinel.Path)
		if sentinel.Dir {
			if err := os.MkdirAll(path, 0755); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, nil, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writeSetupManifest(dir, record); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, setupManifest))
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() > 1024 {
		t.Fatal("marker grew beyond fixed sentinel metadata", info.Size())
	}
	if !completeInstrumentDir(dir, record) {
		t.Fatal("complete sentinels rejected")
	}
	// Other source trees are deliberately not inventoried. Repair on an
	// arbitrary Go error is not automatic; use setup --reset in that case.
	if err := os.MkdirAll(filepath.Join(dir, "other", "tree"), 0755); err != nil {
		t.Fatal(err)
	}
	if !completeInstrumentDir(dir, record) {
		t.Fatal("unrelated source tree affected validation")
	}
	if err := os.RemoveAll(filepath.Join(dir, record.Sentinels[3].Path)); err != nil {
		t.Fatal(err)
	}
	if completeInstrumentDir(dir, record) {
		t.Fatal("missing html/template sentinel accepted")
	}
}

func TestInstrumentLeaseDir(t *testing.T) {
	if instrumentLeaseDir("wrapper", "wrapper/go1.27.2", true, true) != "wrapper" {
		t.Fatal("in-place managed writer must share reader lease")
	}
	if instrumentLeaseDir("revision-a", "original", false, true) != instrumentLeaseDir("revision-b", "original", false, true) {
		t.Fatal("in-place unmanaged writers must share lease across revisions")
	}
	if instrumentLeaseDir("wrapper", "original", false, false) != "wrapper" {
		t.Fatal("managed setup must retain wrapper lease")
	}
}

func TestStagedSetupWaitsForSourceWriter(t *testing.T) {
	parent := t.TempDir()
	source := filepath.Join(parent, "source")
	target := filepath.Join(parent, "target")
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	sourceDir, err := sourceGorootLeaseDir(parent, source)
	if err != nil {
		t.Fatal(err)
	}
	writer, err := lockCache(sourceDir, true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.close()
	entered := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		lease, err := prepareInstrumentDir(target, testSetupRecord("r", "file"), false, func(stage string) error {
			sourceLease, err := lockCache(sourceDir, false, true)
			if err != nil {
				return err
			}
			defer sourceLease.close()
			close(entered)
			return os.WriteFile(filepath.Join(stage, "file"), []byte("complete"), 0644)
		})
		if lease != nil {
			lease.close()
		}
		done <- err
	}()
	select {
	case <-entered:
		t.Error("staged setup read source while in-place writer held lease")
	case <-time.After(50 * time.Millisecond):
	}
	writer.close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("staged setup did not resume after writer exited")
	}
}

func TestSourceGorootLeaseReadOnlyAndAliases(t *testing.T) {
	parent := t.TempDir()
	sourceParent := filepath.Join(parent, "readonly")
	source := filepath.Join(sourceParent, "go")
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "alias")
	if err := os.Symlink(source, alias); err != nil {
		t.Skip(err)
	}
	if err := os.Chmod(sourceParent, 0555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(sourceParent, 0755)
	home := filepath.Join(parent, "home")
	dir, err := sourceGorootLeaseDir(home, source)
	if err != nil {
		t.Fatal(err)
	}
	aliasDir, err := sourceGorootLeaseDir(home, alias)
	if err != nil {
		t.Fatal(err)
	}
	if dir != aliasDir {
		t.Fatal("source aliases have different leases")
	}
	lease, err := lockCache(dir, false, true)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.close()
	writer, err := lockCache(aliasDir, true, false)
	if err == nil {
		writer.close()
		t.Fatal("aliased writer bypassed source reader")
	}
	entries, err := os.ReadDir(sourceParent)
	if err != nil || len(entries) != 1 {
		t.Fatal("source parent was mutated", err)
	}
}
