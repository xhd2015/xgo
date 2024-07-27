//go:build unix
// +build unix

package persistent_after_build

import (
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/cmd"
)

func TestPersistentAfterBuild(t *testing.T) {
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

	args = append(args, "test", "-c", "-o", "test.bin", "--strace", "--strace-dir", "/tmp")

	// build
	err := cmd.Debug().Dir("./testdata").Run(xgoCmd, args...)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Debug().Dir("./testdata").Run("./test.bin", "-test.v")
	if err != nil {
		t.Error(err)
		return
	}
}
