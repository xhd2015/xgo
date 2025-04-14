package overlay_build_cache_error

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/build/util"
	"github.com/xhd2015/xgo/support/cmd"
)

func TestWithinModuleShouldNonOverlayFirstShouldSuccess(t *testing.T) {
	overlayFile, err := setupModulePkg("within_module_first")
	if err != nil {
		t.Fatalf("failed to setup pkg: %v", err)
	}

	gocache, err := getTmpGocache()
	if err != nil {
		t.Fatalf("failed to get tmp gocache: %v", err)
	}
	defer os.RemoveAll(gocache)

	var normOut bytes.Buffer
	err = cmd.Debug().Stdout(&normOut).Stderr(&normOut).Env([]string{"GOCACHE=" + gocache, "GO_BYPASS_XGO=true"}).Run("go", "test", "-v", "./overlay_test_within_module")
	if err != nil {
		t.Errorf("failed to run test: %v", err)
		t.Log(normOut.String())
		return
	}

	var overlayOut bytes.Buffer
	err = cmd.Debug().Stdout(&overlayOut).Stderr(&overlayOut).Env([]string{"GOCACHE=" + gocache, "GO_BYPASS_XGO=true"}).Run("go", "test", "-v", "-overlay", overlayFile, "./overlay_test_within_module")
	if err != nil {
		t.Errorf("failed to run test: %v", err)
		t.Log(overlayOut.String())
		return
	}
}

func TestWithinModuleShouldNonOverlayLaterShouldAlsoSucceed(t *testing.T) {
	overlayFile, err := setupModulePkg("within_module_later")
	if err != nil {
		t.Fatalf("failed to setup reverse: %v", err)
	}

	gocache, err := getTmpGocache()
	if err != nil {
		t.Fatalf("failed to get tmp gocache: %v", err)
	}
	defer os.RemoveAll(gocache)

	var overlayOut bytes.Buffer
	err = cmd.Debug().Stdout(&overlayOut).Stderr(&overlayOut).Env([]string{"GOCACHE=" + gocache, "GO_BYPASS_XGO=true"}).Run("go", "test", "-v", "-overlay", overlayFile, "./overlay_test_within_module")
	if err != nil {
		t.Errorf("failed to run test: %v", err)
		t.Log(overlayOut.String())
		return
	}
	var normOut bytes.Buffer
	err = cmd.Debug().Stdout(&normOut).Stderr(&normOut).Env([]string{"GOCACHE=" + gocache, "GO_BYPASS_XGO=true"}).Run("go", "test", "-v", "./overlay_test_within_module")
	if err != nil {
		t.Errorf("failed to run test: %v", err)
		t.Log(normOut.String())
		return
	}
}

func setupModulePkg(name string) (string, error) {
	overlayFile := "overlay_" + name + ".json"
	err := util.GenerateOverlay("", overlayFile, "./pkg", "pkg.go", "overlay/pkg.go")
	if err != nil {
		return "", fmt.Errorf("failed to generate overlay: %v", err)
	}
	return overlayFile, nil
}
