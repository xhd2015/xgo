package main

import (
	"path/filepath"

	"github.com/xhd2015/xgo/instrument/patch"
)

const interceptCompile = `
   var dlvDebug bool
   xgoDebugPkg := os.Getenv("XGO_HELPER_DEBUG_PKG")
	if xgoDebugPkg!="" {
		var pkg string
		for i, arg := range cmdline {
			if arg == "-p" {
				if i+1 < len(cmdline) {
					pkg = cmdline[i+1]
					break
				}
			}
		}
		// fmt.Printf("debug pkg: %s\n",os.Getenv("XGO_HELPER_DEBUG_PKG"))
		if pkg == xgoDebugPkg {
			dlvDebug = true
			// my compile
			base := filepath.Base(cmdline[0])
			if base == "compile" {
				cmdline[0] = filepath.Join(filepath.Dir(cmdline[0]), "compile.debug")
				oldCmdline := cmdline
				cmdline = []string{
					"dlv",
					"exec",
					"--listen=:2346",
					"--api-version=2",
					"--check-go-version=false",
					"--headless",
					"--",
				}
				cmdline = append(cmdline, oldCmdline...)
			}
		}
	}
`

const pipeoutput = `
	if dlvDebug {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, os.Stdout)
		cmd.Stderr = io.MultiWriter(cmd.Stderr, os.Stderr)
	}
`

func instrumentGc(goroot string) error {
	shellGo := filepath.Join(goroot, "src", "cmd", "go", "internal", "work", "shell.go")

	err := patch.EditFile(shellGo, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin intercept_compile>*/",
			"/*<end intercept_compile>*/",
			[]string{
				"\nfunc (sh *Shell) runOut(dir string",
				"cmdline :=",
				"\n",
			},
			2,
			patch.UpdatePosition_After,
			interceptCompile,
		)
		content = patch.UpdateContent(content,
			"/*<begin pipeoutput>*/",
			"/*<end pipeoutput>*/",
			[]string{
				"\nfunc (sh *Shell) runOut(dir string",
				"err = cmd.Run()",
			},
			1,
			patch.UpdatePosition_Before,
			pipeoutput,
		)

		return content, nil
	})
	if err != nil {
		return err
	}
	return nil
}
