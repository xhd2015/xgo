package test_explorer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	debug_util "github.com/xhd2015/xgo/support/debug"
	"github.com/xhd2015/xgo/support/netutil"
)

type DebugRequest struct {
	Item *TestingItem `json:"item"`
	Path []string     `json:"path"`
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

// deprecated
func setupDebugHandler(server *http.ServeMux, projectDir string, getTestConfig func() (*TestConfig, error)) {
	setupPollHandler(server, "/debug", projectDir, getTestConfig, debug)
}

// deprecated
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

	err := debugTest(goCmd, projectDir, file, buildFlags, []string{"./" + filepath.Dir(relPath)}, fmt.Sprintf("^%s$", name), stdout, stderr, args, env)
	if err != nil {
		return err
	}
	return nil
}

func debugTest(goCmd string, dir string, file string, buildFlags []string, buildArgs []string, runNames string, stdout io.Writer, stderr io.Writer, args []string, env []string) error {
	if goCmd == "" {
		goCmd = "go"
	}
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
	tmpBin := filepath.Join(tmpDir, binName)

	flags := []string{"test", "-c", "-o", tmpBin, "-gcflags=all=-N -l"}
	flags = append(flags, buildFlags...)
	flags = append(flags, buildArgs...)
	err = cmd.Dir(dir).Debug().Stderr(stderr).Stdout(stdout).Run(goCmd, flags...)
	if err != nil {
		return err
	}
	return netutil.ServePort("localhost", 2345, true, 500*time.Millisecond, func(port int) {
		fmt.Fprintln(stderr, debug_util.FormatDlvPrompt(port))
	}, func(port int) error {
		// dlv exec --api-version=2 --listen=localhost:2345 --accept-multiclient --headless ./debug.bin
		runArgs := append([]string{"-test.v", "-test.run", runNames}, args...)
		return cmd.Dir(filepath.Dir(file)).Debug().Stderr(stderr).Stdout(stdout).
			Env(env).
			Run("dlv", debug_util.FormatDlvArgs(tmpBin, port, runArgs)...)
	})
}
