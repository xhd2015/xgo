package core

import (
	"errors"
	"fmt"
	"os"
)

const VERSION = "1.0.11"
const REVISION = "43756010e13cabfae008c1de9d72f98b946b0a09+1"
const NUMBER = 144

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
		fmt.Fprintf(os.Stderr, "WARNING: xgo toolchain: %v\nnote: this message can be turned off by setting %s=true\n", err, XGO_CHECK_TOOLCHAIN_VERSION)
	}
}

func checkVersion() error {
	// xgoVersion, xgoRevision, xgoNumber := XGO_VERSION, XGO_REVISION, XGO_NUMBER
	// _, _, _ = xgoVersion, xgoRevision, xgoNumber
	if XGO_VERSION == "" {
		return errors.New("failed to detect xgo version, requires xgo >= v1.0.9, consider run 'xgo upgrade'")
	}
	if XGO_VERSION == VERSION {
		// if runtime version is larger, that means
		if XGO_NUMBER < NUMBER {
			return errors.New("newer xgo available, consider run 'xgo upgrade'")
		}
		// only one case is feasible: XGO_NUMBER >= runtime NUMBER
		// because xgo can be compatible with older sdk
		return nil
	}
	var updateCmd string
	if XGO_NUMBER < NUMBER {
		updateCmd = "xgo update"
	} else {
		updateCmd = "go get github.com/xhd2015/xgo/runtime@latest"
	}
	return fmt.Errorf("xgo v%s maybe incompatible with xgo/runtime v%s, consider run '%s'", XGO_VERSION, VERSION, updateCmd)
}
