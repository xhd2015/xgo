package test_explorer

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/netutil"
	"github.com/xhd2015/xgo/support/strutil"
)

type DebugRequest struct {
	Item *TestingItem `json:"item"`
}
type DebugResponse struct {
	ID string `json:"id"`
}

type DebugPollRequest struct {
	ID string `json:"id"`
}

type DebugPollResponse struct {
	Events []*TestingItemEvent `json:"events"`
}
type DebugDestroyRequest struct {
	ID string `json:"id"`
}

func setupDebugHandler(server *http.ServeMux, projectDir string, getTestConfig func() (*TestConfig, error)) {
	setupPollHandler(server, "/debug", projectDir, getTestConfig, debug)
}

func debug(ctx *RunContext) error {
	projectDir := ctx.ProjectDir
	file := ctx.File
	relPath := ctx.RelPath
	name := ctx.Name
	stdout := ctx.Stdout
	stderr := ctx.Stderr
	goCmd := ctx.GoCmd
	buildFlags := ctx.BuildFlags
	args := ctx.Args
	env := ctx.Env

	tmpDir, err := os.MkdirTemp("", "go-test-debug")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	binName := "debug.bin"
	baseName := filepath.Base(file)
	if baseName != "" {
		binName = baseName + "-" + binName
	}

	// TODO: find a way to automatically set breakpoint
	// dlvInitFile := filepath.Join(tmpDir, "dlv-init.txt")
	// err = ioutil.WriteFile(dlvInitFile, []byte(fmt.Sprintf("break %s:%d\n", file, req.Item.Line)), 0755)
	// if err != nil {
	// 	return err
	// }
	relPathDir := filepath.Dir(relPath)
	tmpBin := filepath.Join(tmpDir, binName)

	flags := []string{"test", "-c", "-o", tmpBin, "-gcflags=all=-N -l"}
	flags = append(flags, buildFlags...)
	flags = append(flags, "./"+relPathDir)
	err = cmd.Dir(projectDir).Debug().Stderr(stderr).Stdout(stdout).Run(goCmd, flags...)
	if err != nil {
		return err
	}
	err = netutil.ServePort(2345, true, 500*time.Millisecond, func(port int) {
		// user need to set breakpoint explicitly
		fmt.Fprintf(stderr, "dlv listen on localhost:%d\n", port)
		fmt.Fprintf(stderr, "Debug with IDEs:\n")
		fmt.Fprintf(stderr, "  > VSCode: add the following config to .vscode/launch.json configurations:")
		fmt.Fprintf(stderr, "\n%s\n", strutil.IndentLines(formatVscodeConfig(port), "    "))
		fmt.Fprintf(stderr, "  > GoLand: click Add Configuration > Go Remote > localhost:%d\n", port)
		fmt.Fprintf(stderr, "  > Terminal: dlv connect localhost:%d\n", port)
	}, func(port int) error {
		// dlv exec --api-version=2 --listen=localhost:2345 --accept-multiclient --headless ./debug.bin
		return cmd.Dir(filepath.Dir(file)).Debug().Stderr(stderr).Stdout(stdout).
			Env(env).
			Run("dlv",
				append([]string{
					"exec",
					"--api-version=2",
					"--check-go-version=false",
					// NOTE: --init is ignored if --headless
					// "--init", dlvInitFile,
					"--headless",
					// "--allow-non-terminal-interactive=true",
					fmt.Sprintf("--listen=localhost:%d", port),
					tmpBin, "--", "-test.v", "-test.run", fmt.Sprintf("^%s$", name),
				}, args...)...,
			)
	})
	if err != nil {
		return err
	}
	return nil
}

func formatVscodeConfig(port int) string {
	return fmt.Sprintf(`{
	"configurations": [
		{
			"name": "Debug dlv localhost:%d",
			"type": "go",
			"request": "attach",
			"mode": "remote",
			"port": %d,
			"host": "127.0.0.1"
		}
	}
}`, port, port)
}
