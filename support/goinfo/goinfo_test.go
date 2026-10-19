package goinfo

import (
	"strings"
	"testing"
)

func TestParseGoVersion(t *testing.T) {
	for _, tt := range []struct {
		text  string
		minor int
		patch int
	}{
		{"go version go1.17.13 linux/amd64", 17, 13},
		{"go version go1.27.2 linux/amd64", 27, 2},
		{"go version go1.28rc1 linux/amd64", 28, 0},
		{"go version go1.28.0 linux/amd64", 28, 0},
		{"go version devel go1.28-dfc6f7db Fri Oct 16 18:14:08 2026 -0700 linux/amd64", 28, 0},
		{"go version go1.28-devel_dfc6f7db Fri Oct 16 18:14:08 2026 -0700 linux/amd64", 28, 0},
	} {
		t.Run(tt.text, func(t *testing.T) {
			v, err := ParseGoVersion(tt.text)
			if err != nil {
				t.Fatal(err)
			}
			wantDevelopment := strings.Contains(tt.text, "devel")
			if v.Development != wantDevelopment || v.Major != 1 || v.Minor != tt.minor || v.Patch != tt.patch || v.OS != "linux" || v.Arch != "amd64" {
				t.Fatalf("unexpected version: %+v", v)
			}
		})
	}
}
