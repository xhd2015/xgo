package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

const corePkg = "github.com/xhd2015/xgo/runtime/test/core"

func TestCheckVersion(t *testing.T) {
	t.Skip("constant patching is prohibited since xgo v1.1.0")
	tests := []struct {
		xgo     VersionInfo
		runtime VersionInfo
		err     string
	}{
		{
			VersionInfo{},
			VersionInfo{},
			"failed to detect xgo version",
		},
		{
			// good to go if xgo > runtime
			VersionInfo{Version: "1.0.1", Revision: "B", Number: 101},
			VersionInfo{Version: "1.0.0", Revision: "A", Number: 100},
			"xgo/runtime v1.0.0 can be upgraded to v1.0.1",
		},
		{
			// not good to go if xgo < runtime
			VersionInfo{Version: "1.0.1", Revision: "B", Number: 101},
			VersionInfo{Version: "1.0.2", Revision: "C", Number: 102},
			"xgo v1.0.1 maybe incompatible with xgo/runtime v1.0.2",
		},
	}
	for i, tt := range tests {
		tt := tt
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			err := testVersion(tt.xgo, tt.runtime)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if (errMsg == "") != (tt.err == "") {
				t.Fatalf("expect err msg: %q, actual: %q", tt.err, errMsg)
			}
			if !strings.Contains(errMsg, tt.err) {
				t.Fatalf("expect err: %q, actual: %q", tt.err, errMsg)
			}
		})
	}
}

type VersionInfo struct {
	Version  string
	Revision string
	Number   int
}

func testVersion(xgo VersionInfo, runtime VersionInfo) error {
	mock.PatchByName(corePkg, "VERSION", func() string {
		return runtime.Version
	})
	mock.PatchByName(corePkg, "REVISION", func() string {
		return runtime.Revision
	})
	mock.PatchByName(corePkg, "NUMBER", func() int {
		return runtime.Number
	})
	mock.PatchByName(corePkg, "XGO_VERSION", func() string {
		return xgo.Version
	})
	mock.PatchByName(corePkg, "XGO_REVISION", func() string {
		return xgo.Revision
	})
	mock.PatchByName(corePkg, "XGO_NUMBER", func() int {
		return xgo.Number
	})
	return checkVersion()
}
