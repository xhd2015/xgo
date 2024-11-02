package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/xhd2015/xgo/support/cmd"
)

// doc:
//   https://git-scm.com/docs/git-describe
// stackoverflow:
//   https://stackoverflow.com/questions/1474115/how-to-find-the-tag-associated-with-a-given-git-commit

// git describe
func DescribeTag(dir string, ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("requires ref")
	}
	var stderr bytes.Buffer
	// --exact-match
	output, err := cmd.Dir(dir).Stderr(io.MultiWriter(os.Stderr, &stderr)).Output("git", "describe", "--exact-match", "--tags", ref)
	if err != nil {
		// fatal: no tag exactly matches 'bcdddc725ad5b701149b78b2cdbb9a61e72bb21e'
		if err, ok := err.(*exec.ExitError); ok {
			if err.ExitCode() == 128 && bytes.Contains(stderr.Bytes(), []byte("no tag exactly matches")) {
				return "", nil
			}
			return "", fmt.Errorf(stderr.String())
		}
		return "", err
	}
	return output, nil
}
