package trace_dir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/cmd"
)

func TestTraceDir(t *testing.T) {
	var xgoCmd string
	var args []string

	testCmd := os.Getenv("XGO_TEST_COMMAND")
	if testCmd != "" {
		cmds := strings.Split(testCmd, " ")
		xgoCmd = cmds[0]
		args = cmds[1:]
	} else {
		xgoCmd = "xgo"
	}

	tmpDir, err := os.MkdirTemp("", "trace")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(tmpDir)

	args = append(args, "test", "--strace", "--strace-dir", tmpDir)

	// build
	err = cmd.Debug().Dir("./testdata").Run(xgoCmd, args...)
	if err != nil {
		t.Error(err)
		return
	}

	wantExistFile := filepath.Join(tmpDir, "TestGreet.json")
	_, err = os.Stat(wantExistFile)
	if err != nil {
		t.Error(err)
		return
	}
}
