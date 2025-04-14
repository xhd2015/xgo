package overlay_build_cache_error

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/build/util"
	"github.com/xhd2015/xgo/support/cmd"
)

func TestGoModNonOverlayFirstShouldError(t *testing.T) {
	overlayFile, err := setupReverse("gomod_first")
	if err != nil {
		t.Fatalf("failed to setup reverse: %v", err)
	}

	gocache, err := getTmpGocache()
	if err != nil {
		t.Fatalf("failed to get tmp gocache: %v", err)
	}
	defer os.RemoveAll(gocache)
	var normOut bytes.Buffer
	err = cmd.Debug().Stdout(&normOut).Stderr(&normOut).Env([]string{"GOCACHE=" + gocache, "GO_BYPASS_XGO=true"}).Run("go", "test", "-v", "./overlay_test_with_gomod")
	if err != nil {
		t.Log(normOut.String())
		t.Fatalf("failed to run test: %v", err)
	}

	// this should error
	var errOut bytes.Buffer
	afterErr := cmd.Debug().Stdout(&errOut).Stderr(&errOut).Env([]string{"GOCACHE=" + gocache, "GO_BYPASS_XGO=true"}).Run("go", "test", "-v", "-overlay", overlayFile, "./overlay_test_with_gomod")
	if afterErr == nil {
		t.Errorf("expect cache+overlay combination error, actual nil")
	}
	expectContains := "could not import runtime (open : no such file or directory)"
	errOutStr := errOut.String()
	if !strings.Contains(errOutStr, expectContains) {
		t.Errorf("expect containing %q, actual none", expectContains)
		t.Logf("DEBUG: %s", errOutStr)
		return
	}
}

func TestGoModNonOverlayLaterShouldSucceed(t *testing.T) {
	overlayFile, err := setupReverse("gomod_later")
	if err != nil {
		t.Fatalf("failed to setup reverse: %v", err)
	}

	gocache, err := getTmpGocache()
	if err != nil {
		t.Fatalf("failed to get tmp gocache: %v", err)
	}
	defer os.RemoveAll(gocache)

	var overlayOut bytes.Buffer
	err = cmd.Debug().Stdout(&overlayOut).Stderr(&overlayOut).Env([]string{"GOCACHE=" + gocache, "GO_BYPASS_XGO=true"}).Run("go", "test", "-v", "-overlay", overlayFile, "./overlay_test_with_gomod")
	if err != nil {
		t.Log(overlayOut.String())
		t.Fatalf("failed to run test: %v", err)
	}
	// this should error
	var normOut bytes.Buffer
	err = cmd.Debug().Stdout(&normOut).Stderr(&normOut).Env([]string{"GOCACHE=" + gocache, "GO_BYPASS_XGO=true"}).Run("go", "test", "-v", "./overlay_test_with_gomod")
	if err != nil {
		t.Errorf("failed to run test: %v", err)
		t.Log(normOut.String())
		return
	}
}

func getTmpGocache() (string, error) {
	xgoGen := filepath.Join(".xgo", "gen")
	err := os.MkdirAll(xgoGen, 0755)
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp(xgoGen, "go-build")
	if err != nil {
		return "", err
	}
	return filepath.Abs(dir)
}

func setupReverse(name string) (string, error) {
	overlayFile := "overlay_" + name + ".json"
	err := util.GenerateOverlay("", overlayFile, "golang.org/x/example/hello/reverse", "reverse.go", "overlay/reverse/reverse.go")
	if err != nil {
		return "", fmt.Errorf("failed to generate overlay: %v", err)
	}
	return overlayFile, nil
}
