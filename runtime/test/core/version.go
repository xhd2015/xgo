package core

import (
	"errors"
	"fmt"
)

// copied from core/version.go

const VERSION = "1.0.26"
const REVISION = "4e6a5615d778b8909e3315a2ead323822581dd0e+1"
const NUMBER = 198

// these fields will be filled by compiler
const XGO_VERSION = ""
const XGO_REVISION = ""
const XGO_NUMBER = 0

// copy from core

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
