package xgo_integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/osinfo"
)

func checkedShadow(t *testing.T) string {
	checkXgo(t)

	shadow, err := cmd.Output("xgo", "shadow")
	if err != nil {
		t.Fatal(err)
	}
	if shadow == "" {
		t.Fatalf("expect shadow path printed, actual nothing")
	}
	return shadow
}
func TestShadowPrintPath(t *testing.T) {
	shadow := checkedShadow(t)
	shadowGo := filepath.Join(shadow, "go"+osinfo.EXE_SUFFIX)
	stat, err := os.Stat(shadowGo)
	if err != nil {
		t.Fatal(err)
	}
	if stat.IsDir() {
		t.Fatalf("expect shadow go to be file, actual: dir, %s", shadowGo)
	}
}

func TestShadowPathLookupGo(t *testing.T) {
	shadow := checkedShadow(t)
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}

	shadowGo := filepath.Join(shadow, "go"+osinfo.EXE_SUFFIX)
	if goBinary == shadowGo {
		t.Skipf("skipped because shadow on PATH")
	}
	goroot, err := cmd.Output(goBinary, "env", "GOROOT")
	if err != nil {
		t.Fatal(err)
	}

	gorootGo := filepath.Join(goroot, "bin", "go"+osinfo.EXE_SUFFIX)
	if gorootGo != goBinary {
		t.Fatalf("go not inside GOROOT: %s", goBinary)
	}
	var goAfterShadow string
	oldPath := os.Getenv("PATH")
	withPathEnv(shadow+string(filepath.ListSeparator)+oldPath, func() {
		goAfterShadow, err = exec.LookPath("go")
		if err != nil {
			t.Fatal(err)
		}
	})

	if goAfterShadow != shadowGo {
		t.Fatalf("expect  not inside GOROOT: %s", goBinary)
	}
}

func testShadowMock(t *testing.T, env []string) error {
	shadow := checkedShadow(t)
	oldPath := os.Getenv("PATH")
	var err error
	withPathEnv(shadow+string(filepath.ListSeparator)+oldPath, func() {
		err = cmd.Env(env).Dir("./testdata/mock_simple").Run("go", "test", "-v", "./")

	})
	return err
}
func TestShadowMockSucceeds(t *testing.T) {
	err := testShadowMock(t, nil)
	if err != nil {
		t.Fatalf("expect shadow mock success, actual: %v", err)
	}
}

func TestShadowByPassMockFail(t *testing.T) {
	err := testShadowMock(t, []string{"XGO_SHADOW_BYPASS=true"})
	if err == nil {
		t.Fatalf("expect test fail, actually not")
	}
}
