package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveBuildCacheDir(t *testing.T) {
	def := resolveBuildCacheDir("/tmp/xgo/go-instrument/go1.26.0_abc", "", "ns")
	if def != filepath.Join("/tmp/xgo/go-instrument/go1.26.0_abc", "build-cache") {
		t.Fatalf("default: got %q", def)
	}
	got := resolveBuildCacheDir("/tmp/xgo/go-instrument/go1.26.0_abc", "/var/cache/go-build/xgo", "ns")
	if got != filepath.Join("/var/cache/go-build/xgo", "ns") {
		t.Fatalf("override: got %q", got)
	}
}

func TestSharedBuildCacheNamespaceIsLocationIndependent(t *testing.T) {
	ns := sharedBuildCacheNamespace("go1.26.0", "")
	if !strings.HasPrefix(ns, "go1.26.0_xgo"+VERSION+"_") {
		t.Fatalf("namespace %q lacks go/xgo version prefix", ns)
	}
	if strings.ContainsAny(ns, "/\\+") {
		t.Fatalf("namespace %q contains path or revision separators", ns)
	}
	if sharedBuildCacheNamespace("go1.26.0", "-legacy") == ns {
		t.Fatal("instrument suffix must change the namespace")
	}
	if sharedBuildCacheNamespace("go1.26.1", "") == ns {
		t.Fatal("Go patch version must change the namespace")
	}
	first := resolveBuildCacheDir(filepath.Join("slot-a", "cache"), "shared", ns)
	second := resolveBuildCacheDir(filepath.Join("slot-b", "cache"), "shared", ns)
	if first != second {
		t.Fatalf("shared cache depends on runner location: %q != %q", first, second)
	}
}
