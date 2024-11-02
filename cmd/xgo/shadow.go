package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/osinfo"
)

//go:embed shadow
var shadowFS embed.FS

func handleShadow() error {
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

	err = copyEmbedDir(shadowFS, "shadow", tmpDir)
	if err != nil {
		return err
	}

	err = fileutil.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module github.com/xhd2015/xgo/cmd/xgo/shadow

go 1.14
`))
	if err != nil {
		return err
	}

	err = cmd.Dir(tmpDir).Env(appendNativeBuildEnv(nil)).Run(getNakedGo(), "build", "-o", filepath.Join(shadowDir, "go"+osinfo.EXE_SUFFIX), "./")
	if err != nil {
		return err
	}
	fmt.Println(shadowDir)

	return nil
}
