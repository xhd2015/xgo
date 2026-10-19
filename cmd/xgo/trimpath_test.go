package main

import "testing"

func TestBuildTrimpath(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		buildFlags []string
		goflags    string
		want       bool
	}{
		{name: "default"},
		{name: "command line", args: []string{"-trimpath", "./..."}, want: true},
		{name: "double dash", args: []string{"--trimpath=true"}, want: true},
		{name: "build flags", buildFlags: []string{"-trimpath"}, want: true},
		{name: "GOFLAGS", goflags: "-mod=readonly -trimpath", want: true},
		{name: "explicit false", args: []string{"-trimpath=false"}, goflags: "-trimpath"},
		{name: "last flag wins", args: []string{"-trimpath", "--trimpath=false"}},
		{name: "args override build flags", args: []string{"-trimpath=false"}, buildFlags: []string{"-trimpath"}},
		{name: "build flags override GOFLAGS", buildFlags: []string{"-trimpath=false"}, goflags: "-trimpath"},
		{name: "boolean spelling", args: []string{"-trimpath=1"}, want: true},
		{name: "unrelated flag", args: []string{"-ldflags=-trimpath"}},
		{name: "after separator", args: []string{"--", "-trimpath"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildTrimpath(tt.args, tt.buildFlags, tt.goflags); got != tt.want {
				t.Fatalf("buildTrimpath() = %v, want %v", got, tt.want)
			}
		})
	}
}
