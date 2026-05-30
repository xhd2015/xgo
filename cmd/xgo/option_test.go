package main

import (
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

func TestParseUseFilePatches(t *testing.T) {
	pbTrue := ptrBool(true)
	pbFalse := ptrBool(false)

	tests := []struct {
		name    string
		args    []string
		want    *bool
		wantErr bool
	}{
		{name: "not set", args: []string{"build", "./..."}, want: nil},
		{name: "bare flag", args: []string{"build", "--use-file-patches", "./..."}, want: pbTrue},
		{name: "explicit true", args: []string{"build", "--use-file-patches=true", "./..."}, want: pbTrue},
		{name: "explicit false", args: []string{"build", "--use-file-patches=false", "./..."}, want: pbFalse},
		{name: "invalid value", args: []string{"build", "--use-file-patches=bad", "./..."}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseOptions(tt.args[0], tt.args[1:])
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.want == nil {
				if opts.useFilePatches != nil {
					t.Fatalf("expected nil, got %v", *opts.useFilePatches)
				}
				return
			}
			if opts.useFilePatches == nil {
				t.Fatal("expected non-nil, got nil")
			}
			if *opts.useFilePatches != *tt.want {
				t.Fatalf("expected %v, got %v", *tt.want, *opts.useFilePatches)
			}
		})
	}
}

func TestResolveUseFilePatches(t *testing.T) {
	go124 := &goinfo.GoVersion{Major: 1, Minor: 24}
	go125 := &goinfo.GoVersion{Major: 1, Minor: 25}
	go123 := &goinfo.GoVersion{Major: 1, Minor: 23}
	go126 := &goinfo.GoVersion{Major: 1, Minor: 26}
	go217 := &goinfo.GoVersion{Major: 1, Minor: 17}

	pbTrue := ptrBool(true)
	pbFalse := ptrBool(false)

	tests := []struct {
		name        string
		explicit    *bool
		goVersion   *goinfo.GoVersion
		want        bool
		wantWarning bool
		wantErr     bool
	}{
		// nil explicit — version-based defaults
		{name: "go1.25 default", goVersion: go125, want: true},
		{name: "go1.24 default", goVersion: go124, want: false},
		{name: "go1.23 default", goVersion: go123, want: false},
		{name: "go1.17 default", goVersion: go217, want: false},
		{name: "nil version default", want: false},

		// explicit true — validated
		{name: "go1.25 explicit true", explicit: pbTrue, goVersion: go125, want: true},
		{name: "go1.24 explicit true", explicit: pbTrue, goVersion: go124, want: true},
		{name: "go1.23 explicit true", explicit: pbTrue, goVersion: go123, wantErr: true},
		{name: "go1.26 explicit true", explicit: pbTrue, goVersion: go126, wantErr: true},
		{name: "nil version explicit true", explicit: pbTrue, wantErr: true},

		// explicit false — warning on unsupported, no error
		{name: "go1.25 explicit false", explicit: pbFalse, goVersion: go125, want: false},
		{name: "go1.24 explicit false", explicit: pbFalse, goVersion: go124, want: false},
		{name: "go1.23 explicit false", explicit: pbFalse, goVersion: go123, want: false, wantWarning: true},
		{name: "go1.26 explicit false", explicit: pbFalse, goVersion: go126, want: false, wantWarning: true},
		{name: "nil version explicit false", explicit: pbFalse, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warning, err := resolveUseFilePatches(tt.explicit, tt.goVersion)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if tt.wantWarning && warning == "" {
				t.Fatal("expected warning, got none")
			}
			if !tt.wantWarning && warning != "" {
				t.Fatalf("expected no warning, got: %s", warning)
			}
		})
	}
}

func ptrBool(b bool) *bool { return &b }
