package build

import (
	"runtime"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

func TestNeedExternalLinker(t *testing.T) {
	t.Setenv("GOOS", "")
	t.Setenv("GOARCH", "")

	tests := []struct {
		name               string
		goVersion          *goinfo.GoVersion
		wantOnAppleSilicon bool
	}{
		{"go1.17", &goinfo.GoVersion{Major: 1, Minor: 17}, true},
		{"go1.18", &goinfo.GoVersion{Major: 1, Minor: 18}, true},
		{"go1.21", &goinfo.GoVersion{Major: 1, Minor: 21}, true},
		{"go1.22.0", &goinfo.GoVersion{Major: 1, Minor: 22, Patch: 0}, true},
		{"go1.22.8", &goinfo.GoVersion{Major: 1, Minor: 22, Patch: 8}, true},
		{"go1.22.12", &goinfo.GoVersion{Major: 1, Minor: 22, Patch: 12}, true},
		{"go1.23", &goinfo.GoVersion{Major: 1, Minor: 23}, false},
		{"go1.24", &goinfo.GoVersion{Major: 1, Minor: 24}, false},
		{"go1.25", &goinfo.GoVersion{Major: 1, Minor: 25}, false},
		{"nil version", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedExternalLinker(tt.goVersion)

			want := false
			if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
				want = tt.wantOnAppleSilicon
			}

			if got != want {
				t.Errorf("NeedExternalLinker(%v) = %v, want %v (host %s/%s, target %s/%s)",
					tt.goVersion, got, want, GetHostGOOS(), GetHostGOARCH(), GetTargetGOOS(), GetTargetGOARCH())
			}
		})
	}
}

func TestNeedExternalLinker_crossCompileToNonDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("test only meaningful on darwin/arm64 host")
	}

	t.Setenv("GOOS", "linux")
	gv := &goinfo.GoVersion{Major: 1, Minor: 17}
	if got := NeedExternalLinker(gv); got {
		t.Error("NeedExternalLinker(go1.17) with GOOS=linux should return false")
	}
}

func TestNeedExternalLinker_crossCompileToDarwinAmd64(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("test only meaningful on darwin/arm64 host")
	}

	t.Setenv("GOOS", "")
	t.Setenv("GOARCH", "amd64")
	gv := &goinfo.GoVersion{Major: 1, Minor: 17}
	if got := NeedExternalLinker(gv); !got {
		t.Error("NeedExternalLinker(go1.17) with GOARCH=amd64 on darwin should return true (macOS linker supports amd64)")
	}
}

func TestExternalLinkerFlags(t *testing.T) {
	t.Setenv("GOOS", "")
	t.Setenv("GOARCH", "")

	t.Run("returns flags when external linker needed", func(t *testing.T) {
		gv := &goinfo.GoVersion{Major: 1, Minor: 17}
		got := ExternalLinkerFlags(gv)

		wantLen := 0
		if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
			wantLen = 1
		}

		if len(got) != wantLen {
			t.Errorf("ExternalLinkerFlags() len = %d, want %d", len(got), wantLen)
		}
		if len(got) > 0 && got[0] != "-ldflags=-linkmode=external" {
			t.Errorf("ExternalLinkerFlags()[0] = %q, want %q", got[0], "-ldflags=-linkmode=external")
		}
	})

	t.Run("returns nil when external linker not needed", func(t *testing.T) {
		gv := &goinfo.GoVersion{Major: 1, Minor: 24}
		got := ExternalLinkerFlags(gv)
		if got != nil {
			t.Errorf("ExternalLinkerFlags(go1.24) = %v, want nil", got)
		}
	})

	t.Run("returns nil for nil version", func(t *testing.T) {
		got := ExternalLinkerFlags(nil)
		if got != nil {
			t.Errorf("ExternalLinkerFlags(nil) = %v, want nil", got)
		}
	})
}
