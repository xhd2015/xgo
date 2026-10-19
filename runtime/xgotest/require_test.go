package xgotest

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

func TestRequire(t *testing.T) {
	if core.Instrumented() {
		t.Skip("plain-Go subprocess test; xgo-built behavior is covered by runtime/test/core/instrumented")
	}
	for _, mode := range []string{"", "fail", "skip", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRequireChild$", "-test.v")
			cmd.Env = append(os.Environ(), "XGO_REQUIRE_CHILD=1", "XGO_REQUIRE="+mode)
			out, err := cmd.CombinedOutput()
			if (err == nil) != (mode == "skip") {
				t.Fatalf("exit = %v, output:\n%s", err, out)
			}
			if !strings.Contains(string(out), runtime.NotInstrumentedMessage) || strings.Contains(string(out), "REQUIRE_RETURNED") {
				t.Fatalf("Require did not stop test with expected message:\n%s", out)
			}
			if mode == "skip" && !strings.Contains(string(out), "--- SKIP: TestRequireChild") {
				t.Fatalf("Require did not skip:\n%s", out)
			}
			if mode == "invalid" && !strings.Contains(string(out), "invalid XGO_REQUIRE") {
				t.Fatalf("missing invalid mode diagnostic:\n%s", out)
			}
		})
	}
}

func TestRequireChild(t *testing.T) {
	if os.Getenv("XGO_REQUIRE_CHILD") != "1" {
		return
	}
	Require(t)
	fmt.Println("REQUIRE_RETURNED")
}
