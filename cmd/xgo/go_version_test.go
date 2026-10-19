package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckGoVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a shell script as the go executable")
	}
	for _, tt := range []struct {
		version string
		wantErr bool
	}{
		{"go1.16.15", true},
		{"go1.17.13", false},
		{"go1.27.2", false},
		{"go1.28rc1", false},
		{"go1.28-devel_dfc6f7db", false},
		{"go1.29rc1", true},
		{"go2.0.0", true},
	} {
		t.Run(tt.version, func(t *testing.T) {
			goroot := t.TempDir()
			bin := filepath.Join(goroot, "bin")
			if err := os.Mkdir(bin, 0755); err != nil {
				t.Fatal(err)
			}
			script := fmt.Sprintf("#!/bin/sh\necho 'go version %s linux/amd64'\n", tt.version)
			if err := os.WriteFile(filepath.Join(bin, "go"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			if _, err := checkGoVersion(goroot, true); (err != nil) != tt.wantErr {
				t.Fatalf("checkGoVersion: %v, wantErr=%v", err, tt.wantErr)
			}
			if _, err := checkGoVersion(goroot, false); err != nil {
				t.Fatalf("version gate disabled: %v", err)
			}
		})
	}
}
