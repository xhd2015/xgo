package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/instrument/build"
	instrument_embed "github.com/xhd2015/xgo/instrument/embed"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/osinfo"
)

//go:embed shadow
var shadowFS embed.FS

// Deprecated: use `xgo setup` instead
func handleShawdow() error {
	xgoHome, err := getOrMakeAbsXgoHome("")
	if err != nil {
		return err
	}

	shadowDir := filepath.Join(xgoHome, "shadow")
	err = os.MkdirAll(shadowDir, 0755)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "shadow-build")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	err = instrument_embed.CopyDir(shadowFS, "shadow", tmpDir, instrument_embed.CopyOptions{})
	if err != nil {
		return err
	}

	err = fileutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module github.com/xhd2015/xgo/cmd/xgo/shadow

go 1.14
`))
	if err != nil {
		return err
	}

	err = cmd.Dir(tmpDir).Env(build.AppendNativeBuildEnv(nil)).Run(getNakedGo(), "build", "-o", filepath.Join(shadowDir, "go"+osinfo.EXE_SUFFIX), "./")
	if err != nil {
		return err
	}
	fmt.Println(shadowDir)

	return nil
}
