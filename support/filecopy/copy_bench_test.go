package filecopy_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/filecopy"
)

// conclusion:
//   10 goroutine seems enough to copy(go1.22.1 cost ~1.9s, has 12876 files)

// go test -run TestExampleCopy1g -v ./support/filecopy
func TestExampleCopy1g(t *testing.T) {
	testCopyDir(t, "go1.22.1", 1)
}

// go test -run TestExampleCopy5g -v ./support/filecopy
func TestExampleCopy5g(t *testing.T) {
	testCopyDir(t, "go1.22.1", 5)
}

// go test -run TestExampleCopy10g -v ./support/filecopy
func TestExampleCopy10g(t *testing.T) {
	testCopyDir(t, "go1.22.1", 10)
}

// go test -run TestExampleCopy10g -v ./support/filecopy
func TestExampleCopy20g(t *testing.T) {
	testCopyDir(t, "go1.22.1", 20)
}

// go test -run TestExampleCopy50g -v ./support/filecopy
func TestExampleCopy50g(t *testing.T) {
	testCopyDir(t, "go1.22.1", 50)
}

// go test -run TestExampleCopy100g -v ./support/filecopy
func TestExampleCopy100g(t *testing.T) {
	testCopyDir(t, "go1.22.1", 100)
}

func getGitRoot() (string, error) {
	gitDir, err := cmd.Output("git", "rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Dir(gitDir))
}
func testCopyDir(t *testing.T, src string, n int) {
	root, err := getGitRoot()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp("/tmp", "cp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	// fmt.Printf("copy to %s\n", dir)
	srcDir := filepath.Join(root, "go-release", src)

	_, err = os.Stat(srcDir)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	err = filecopy.NewOptions().Concurrent(n).CopyReplaceDir(srcDir, dir)
	if err != nil {
		t.Fatal(err)
	}
	end := time.Now()

	cost := end.Sub(start)

	t.Logf("copy: %s, n: %d, cost: %v", src, n, cost)
	// copy: go1.22.1, n: 1, cost: 8.070770324s
	// copy: go1.22.1, n: 10, cost: 1.870731672s
	// copy: go1.22.1, n: 50, cost: 1.968898761s
	// copy: go1.22.1, n: 100, cost: 2.015457571s
}
