package core

import (
	"errors"
	"fmt"
	"os"
)

const VERSION = "1.0.26"
const REVISION = "d19c85ac922d26f26038dc99ae3049de0d91d2f6+1"
const NUMBER = 199

// these fields will be filled by compiler
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
		// if runtime version is larger, that means
		if XGO_NUMBER < NUMBER {
			return errors.New("newer xgo available, consider run: xgo upgrade")
		}
		// only one case is feasible: XGO_NUMBER >= runtime NUMBER
		// because xgo can be compatible with older sdk
		return nil
	}
	if XGO_NUMBER == NUMBER {
		// good to go
		return nil
	}
	var msg string
	if XGO_NUMBER < NUMBER {
		updateCmd := "xgo upgrade"
		msg = fmt.Sprintf("xgo v%s maybe incompatible with xgo/runtime v%s, consider run: %s", XGO_VERSION, VERSION, updateCmd)
	} else {
		updateCmd := "go get github.com/xhd2015/xgo/runtime@latest"
		msg = fmt.Sprintf("xgo/runtime v%s can be upgraded to v%s, consider run: %s", VERSION, XGO_VERSION, updateCmd)
	}
	return errors.New(msg)
}
