package core

import (
	"errors"
	"fmt"
	"os"
)

// these fields are automatically copied
// from when running
// `go run ./script/generate runtime/core/version.go`
const VERSION = "1.1.4"
const REVISION = "f4d4eabcec51dc30ae0d478b0b866f9c581a330c+1"
const NUMBER = 428

// these fields will be filled by compiler when xgo builds with xgo/runtime
// see CORE_VERSION in cmd/xgo/version.go
const XGO_VERSION = ""
const XGO_REVISION = ""
const XGO_NUMBER = 0

const XGO_CHECK_TOOLCHAIN_VERSION = "XGO_CHECK_TOOLCHAIN_VERSION"

func init() {
	envVal := os.Getenv(XGO_CHECK_TOOLCHAIN_VERSION)
	if envVal == "false" || envVal == "off" {
		return
	}
	err := checkVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: xgo toolchain: %v\nnote: this message can be turned off by setting %s=false\n", err, XGO_CHECK_TOOLCHAIN_VERSION)
	}
}
func checkVersion() error {
	// xgoVersion, xgoRevision, xgoNumber := XGO_VERSION, XGO_REVISION, XGO_NUMBER
	// _, _, _ = xgoVersion, xgoRevision, xgoNumber
	if XGO_VERSION == "" {
		return errors.New("failed to detect xgo version, consider install xgo: go install github.com/xhd2015/xgo/cmd/xgo@latest")
	}
	if XGO_VERSION == VERSION {
		// if xgo version same with runtime version, then
		// different numbers are compatible with each other,
		// e.g.  xgo v1.0.29 <-> runtime v1.0.29
		//       [ok] xgo 205, runtime 205
		//       [ok] xgo 205, runtime 206
		//       [ok] xgo 205, runtime 204
		return nil
	}
	if XGO_NUMBER == NUMBER {
		// good to go
		return nil
	}
	var msg string
	if XGO_NUMBER < NUMBER {
		// incompatible possibly: xgo < runtime
		updateCmd := "xgo upgrade"
		msg = fmt.Sprintf("xgo v%s maybe incompatible with xgo/runtime v%s, consider run: %s", XGO_VERSION, VERSION, updateCmd)
	} else {
		// compatible: xgo >= runtime
		updateCmd := "go get github.com/xhd2015/xgo/runtime@latest"
		msg = fmt.Sprintf("xgo/runtime v%s can be upgraded to v%s, consider run: %s", VERSION, XGO_VERSION, updateCmd)
	}
	return errors.New(msg)
}
