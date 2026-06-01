package build

import (
	"runtime"
	"testing"
)

func TestGetHostGOOS(t *testing.T) {
	got := GetHostGOOS()
	if got != runtime.GOOS {
		t.Errorf("GetHostGOOS() = %q, want %q", got, runtime.GOOS)
	}
}

func TestGetHostGOARCH(t *testing.T) {
	got := GetHostGOARCH()
	if got != runtime.GOARCH {
		t.Errorf("GetHostGOARCH() = %q, want %q", got, runtime.GOARCH)
	}
}

func TestGetTargetGOOS_fallsBackToHost(t *testing.T) {
	t.Setenv("GOOS", "")
	got := GetTargetGOOS()
	want := runtime.GOOS
	if got != want {
		t.Errorf("GetTargetGOOS() = %q, want %q (host)", got, want)
	}
}

func TestGetTargetGOARCH_fallsBackToHost(t *testing.T) {
	t.Setenv("GOARCH", "")
	got := GetTargetGOARCH()
	want := runtime.GOARCH
	if got != want {
		t.Errorf("GetTargetGOARCH() = %q, want %q (host)", got, want)
	}
}

func TestGetTargetGOOS_returnsEnvWhenSet(t *testing.T) {
	t.Setenv("GOOS", "linux")
	got := GetTargetGOOS()
	if got != "linux" {
		t.Errorf("GetTargetGOOS() = %q, want %q", got, "linux")
	}
}

func TestGetTargetGOARCH_returnsEnvWhenSet(t *testing.T) {
	t.Setenv("GOARCH", "amd64")
	got := GetTargetGOARCH()
	if got != "amd64" {
		t.Errorf("GetTargetGOARCH() = %q, want %q", got, "amd64")
	}
}
