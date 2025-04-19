package instrument

import (
	"fmt"
	"path/filepath"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

const XGO_HELPER_DEBUG_PKG = "XGO_HELPER_DEBUG_PKG"

const VScodeDebug = `{
    "configurations": [
        {
            "name": "Debug dlv localhost:2346",
            "type": "go",
            "debugAdapter": "dlv-dap",
            "request": "attach",
            "mode": "remote",
            "port": 2346,
            "host": "127.0.0.1",
            "cwd": "./"
        }
    ]
}`

const interceptCompile = `
   var dlvDebug bool
   xgoDebugPkg := os.Getenv("` + XGO_HELPER_DEBUG_PKG + `")
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
		// fmt.Printf("debug pkg: %s\n",os.Getenv("` + XGO_HELPER_DEBUG_PKG + `"))
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
		fmt.Fprintf(os.Stderr, "VSCode: add the following config to .vscode/launch.json:\n%s\nThen set breakpoint at %s\n", ` + "`" + VScodeDebug + "`" + `, runtime.GOROOT() + "/src/cmd/compile/main.go")
	}
`

func InstrumentGc(goroot string, goVersion *goinfo.GoVersion) error {
	// go1.22: src/cmd/go/internal/work/shell.go
	// go1.21: src/cmd/go/internal/work/exec.go
	if goVersion.Major != 1 {
		return fmt.Errorf("go version %s is not supported", goVersion.String())
	}
	var shellGo string
	var fnAnchor string
	var runAnchor string
	if goVersion.Minor >= 22 {
		shellGo = filepath.Join(goroot, "src", "cmd", "go", "internal", "work", "shell.go")
		fnAnchor = "\nfunc (sh *Shell) runOut(dir string"
		runAnchor = "err = cmd.Run()"
	} else {
		shellGo = filepath.Join(goroot, "src", "cmd", "go", "internal", "work", "exec.go")
		fnAnchor = `func (b *Builder) runOut(a *Action, dir string,`
		runAnchor = "err := cmd.Run()"
	}

	err := patch.EditFile(shellGo, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin intercept_compile>*/",
			"/*<end intercept_compile>*/",
			[]string{
				fnAnchor,
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
				fnAnchor,
				runAnchor,
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
