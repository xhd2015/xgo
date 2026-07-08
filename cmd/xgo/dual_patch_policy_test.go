package main

import (
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

// TestDualPatchPolicy documents the go1.24–go1.27 patch routing policy:
//   - go1.24–go1.25: legacy instrument_go or file-based (explicit choice)
//   - go1.26: dual-path — file-based by default, legacy via --use-file-patches=false
//   - go1.27+: file-based only
func TestDualPatchPolicy(t *testing.T) {
	versions := []struct {
		name string
		ver  *goinfo.GoVersion
	}{
		{"go1.24", &goinfo.GoVersion{Major: 1, Minor: 24}},
		{"go1.25", &goinfo.GoVersion{Major: 1, Minor: 25}},
		{"go1.26", &goinfo.GoVersion{Major: 1, Minor: 26}},
		{"go1.27", &goinfo.GoVersion{Major: 1, Minor: 27}},
	}

	type want struct {
		value bool
		err   bool
	}

	// nil explicit → version-based default
	defaults := map[string]want{
		"go1.24": {value: false},
		"go1.25": {value: true},
		"go1.26": {value: true},
		"go1.27": {value: true},
	}

	// explicit true → file-based where supported (go1.24+)
	explicitTrue := map[string]want{
		"go1.24": {value: true},
		"go1.25": {value: true},
		"go1.26": {value: true},
		"go1.27": {value: true},
	}

	// explicit false → legacy where allowed; go1.27+ rejects
	explicitFalse := map[string]want{
		"go1.24": {value: false},
		"go1.25": {value: false},
		"go1.26": {value: false},
		"go1.27": {err: true},
	}

	for _, mode := range []struct {
		name    string
		explicit func(bool) *bool
		wants    map[string]want
	}{
		{"default", func(b bool) *bool { return nil }, defaults},
		{"explicit true", ptrBool, explicitTrue},
		{"explicit false", func(b bool) *bool { v := false; return &v }, explicitFalse},
	} {
		for _, v := range versions {
			t.Run(mode.name+"/"+v.name, func(t *testing.T) {
				var explicit *bool
				if mode.explicit != nil {
					explicit = mode.explicit(true)
				}
				want := mode.wants[v.name]
				got, _, err := resolveUseFilePatches(explicit, v.ver)
				if want.err {
					if err == nil {
						t.Fatal("expected error, got nil")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if got != want.value {
					t.Fatalf("expected useFilePatches=%v, got %v", want.value, got)
				}
			})
		}
	}
}