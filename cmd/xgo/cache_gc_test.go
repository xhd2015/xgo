package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGCStaleCaches(t *testing.T) {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	parent := t.TempDir()
	mk := func(name string, mark time.Time, modTime time.Time) string {
		dir := filepath.Join(parent, name)
		if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
			t.Fatal(err)
		}
		if !mark.IsZero() {
			markFile := filepath.Join(dir, gcMarkFile)
			if err := os.WriteFile(markFile, nil, 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(markFile, mark, mark); err != nil {
				t.Fatal(err)
			}
		}
		for _, p := range []string{filepath.Join(dir, "sub"), dir} {
			if err := os.Chtimes(p, modTime, modTime); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	old := now.Add(-staleBuildCacheAge - time.Hour)
	recent := now.Add(-time.Hour)

	inUse := mk("go1.27.2_in_use", time.Time{}, old)
	staleMarked := mk("go1.26.1_stale", old, old)
	recentMarked := mk("go1.26.2_recent", recent, old)
	staleUnmarked := mk("go1.25.0_unmarked", time.Time{}, old)
	recentUnmarked := mk("go1.25.1_unmarked_recent", time.Time{}, recent)
	foreign := mk("other", old, old)

	gcStaleCaches([]gcUnit{{dir: inUse, namePrefix: "go", maxAge: staleBuildCacheAge}}, now)

	for dir, wantExist := range map[string]bool{
		inUse:          true,
		staleMarked:    false,
		recentMarked:   true,
		staleUnmarked:  false,
		recentUnmarked: true,
		foreign:        true,
	} {
		_, err := os.Stat(dir)
		if exist := err == nil; exist != wantExist {
			t.Errorf("%s: exist=%v, want %v", filepath.Base(dir), exist, wantExist)
		}
	}
	if info, err := os.Stat(filepath.Join(inUse, gcMarkFile)); err != nil || !info.ModTime().Equal(now) {
		t.Errorf("in-use dir not marked: %v", err)
	}

	// scanned at most once per interval
	again := mk("go1.24.0_stale", old, old)
	gcStaleCaches([]gcUnit{{dir: inUse, namePrefix: "go", maxAge: staleBuildCacheAge}}, now.Add(time.Hour))
	if _, err := os.Stat(again); err != nil {
		t.Errorf("rescanned within gc interval: %v", err)
	}
	gcStaleCaches([]gcUnit{{dir: inUse, namePrefix: "go", maxAge: staleBuildCacheAge}}, now.Add(gcInterval+time.Hour))
	if _, err := os.Stat(again); err == nil {
		t.Errorf("stale dir kept after gc interval")
	}
}

func TestCacheGCUnitsSkipsExternalInstrumentDir(t *testing.T) {
	units := cacheGCUnits("/goroot-parent", "/tmp/xgo/go-instrument/x", "/tmp/xgo/go-instrument/x/build-cache", "", false)
	if len(units) != 1 || units[0].dir != "/tmp/xgo/go-instrument/x" || units[0].noMark {
		t.Fatalf("unexpected units: %+v", units)
	}
	units = cacheGCUnits("/home/.xgo/go-instrument/y", "/tmp/xgo/go-instrument/x", "/shared/go1.27.2_xgo1.2.9_abc", "/shared", true)
	if len(units) != 3 ||
		units[0].dir != "/shared/go1.27.2_xgo1.2.9_abc" || units[0].nameContains != "_xgo" ||
		units[1].dir != "/tmp/xgo/go-instrument/x" || !units[1].noMark ||
		units[2].dir != "/home/.xgo/go-instrument/y" || units[2].cacheParent != filepath.Dir("/tmp/xgo/go-instrument/x") {
		t.Fatalf("unexpected units: %+v", units)
	}
}

// An instrumented GOROOT written by an xgo that did not mark it is kept
// while its default build cache is in use, and otherwise only removed after
// staleUnmarkedInstrumentDirAge.
func TestGCUnmarkedInstrumentDir(t *testing.T) {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	home := t.TempDir()
	cacheParent := t.TempDir()
	setTimes := func(dir string, at time.Time) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(dir, at, at); err != nil {
			t.Fatal(err)
		}
	}
	created := now.Add(-staleInstrumentDirAge - time.Hour)
	inUse := filepath.Join(home, "go1.27.2_current")
	setTimes(inUse, created)
	withCache := filepath.Join(home, "go1.26.1_with_cache")
	setTimes(withCache, created)
	cacheDir, ok := instrumentCacheDirOf(cacheParent, withCache)
	if !ok {
		t.Fatal("no cache dir")
	}
	setTimes(cacheDir, now.Add(-time.Hour))
	noCache := filepath.Join(home, "go1.25.1_no_cache")
	setTimes(noCache, created)
	veryOld := filepath.Join(home, "go1.24.1_very_old")
	setTimes(veryOld, now.Add(-staleUnmarkedInstrumentDirAge-time.Hour))
	tomb := filepath.Join(home, gcTombPrefix+"go1.23.0_x")
	setTimes(tomb, created)

	gcStaleCaches([]gcUnit{{dir: inUse, namePrefix: "go", maxAge: staleInstrumentDirAge, cacheParent: cacheParent}}, now)
	for dir, wantExist := range map[string]bool{
		inUse:     true,
		withCache: true,
		noCache:   true,
		veryOld:   false,
		tomb:      false,
	} {
		_, err := os.Stat(dir)
		if exist := err == nil; exist != wantExist {
			t.Errorf("%s: exist=%v, want %v", filepath.Base(dir), exist, wantExist)
		}
	}
}

func TestGCNoMarkUnitIsNotMarked(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "go1.27.2_x")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	gcStaleCaches([]gcUnit{{dir: dir, namePrefix: "go", maxAge: staleBuildCacheAge, noMark: true}}, time.Now())
	if _, err := os.Stat(filepath.Join(dir, gcMarkFile)); err == nil {
		t.Errorf("noMark unit was marked")
	}
}

func TestGCCapsBuildCaches(t *testing.T) {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	parent := t.TempDir()
	mk := func(name string, mark time.Time) string {
		dir := filepath.Join(parent, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if !mark.IsZero() {
			markFile := filepath.Join(dir, gcMarkFile)
			if err := os.WriteFile(markFile, nil, 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(markFile, mark, mark); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	// a daily scan just ran: only a fresh in-use cache triggers another
	recordGC(parent, now.Add(-time.Minute))

	inUse := mk("go1.27.2_new", time.Time{})
	ago := func(d time.Duration) time.Time { return now.Add(-d) }
	busy := mk("go1.27.2_busy", ago(10*time.Minute))
	h2 := mk("go1.27.2_h2", ago(2*time.Hour))
	h3 := mk("go1.27.2_h3", ago(3*time.Hour))
	h4 := mk("go1.27.2_h4", ago(4*time.Hour))
	foreign := mk("other", ago(4*time.Hour))

	gcStaleCaches([]gcUnit{{dir: inUse, namePrefix: "go", maxAge: staleBuildCacheAge, maxKeep: 3}}, now)

	for dir, wantExist := range map[string]bool{
		inUse:   true,
		busy:    true,
		h2:      true,
		h3:      false,
		h4:      false,
		foreign: true,
	} {
		_, err := os.Stat(dir)
		if exist := err == nil; exist != wantExist {
			t.Errorf("%s: exist=%v, want %v", filepath.Base(dir), exist, wantExist)
		}
	}
}

func TestGCCapSparesRecentlyUsedCaches(t *testing.T) {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	parent := t.TempDir()
	var dirs []string
	for i, d := range []time.Duration{0, 5, 10, 20, 30} {
		dir := filepath.Join(parent, fmt.Sprintf("go1.27.2_%d", i))
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		mark := filepath.Join(dir, gcMarkFile)
		if err := os.WriteFile(mark, nil, 0644); err != nil {
			t.Fatal(err)
		}
		at := now.Add(-d * time.Minute)
		if err := os.Chtimes(mark, at, at); err != nil {
			t.Fatal(err)
		}
		dirs = append(dirs, dir)
	}
	gcStaleCaches([]gcUnit{{dir: dirs[0], namePrefix: "go", maxAge: staleBuildCacheAge, maxKeep: 3}}, now)
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("%s used within minEvictAge was removed", filepath.Base(dir))
		}
	}
}
