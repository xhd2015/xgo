package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

var ErrNotExists = errors.New("reference does exist")

func RevParse(dir string, ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("requires revision")
	}
	return cmd.Dir(dir).Output("git", "rev-parse", ref+"^{commit}")
}

func RevParseVerified(dir string, ref string) (string, error) {
	commit, err := revParseVerified(dir, ref)
	if err != nil {
		if err == ErrNotExists {
			return "", fmt.Errorf("%s does not exist or has been deleted", trimRef(ref))
		}
		return "", err
	}
	return commit, nil
}

func RevParseOrEmpty(dir string, ref string) (string, error) {
	commit, err := revParseVerified(dir, ref)
	if err != nil {
		if err == ErrNotExists {
			return "", nil
		}
		return "", err
	}
	return commit, nil
}

func revParseVerified(dir string, ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("requires revision")
	}
	// --verify: verify exactly one parameter is provided
	// --quiet: when --verify, exit non-zero if error
	// S^{commit}: return commit id if it is an annotated tag, see https://git-scm.com/docs/git-rev-parse
	commit, err := cmd.Dir(dir).Output("git", "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return "", ErrNotExists

		}
		return "", err
	}
	return commit, nil
}

func convertRefError(dir string, ref string, err error) error {
	_, revErr := revParseVerified(dir, ref)
	if revErr == ErrNotExists {
		return fmt.Errorf("%s does not exist or has been deleted", trimRef(ref))
	}
	return err
}

func trimRef(ref string) string {
	return strings.TrimPrefix(ref, "origin/")
}
