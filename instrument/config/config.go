package config

import (
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/support/goinfo"
)

var PREDEFINED_STD_PKGS = []string{
	"time",
	"os",
	"os/exec",
	"net",
	"net/http",
	"io",
	"io/ioutil",
}

type PkgConfig struct {
	WhitelistFunc       map[string]bool
	WhitelistFuncPrefix []string
}

func CheckInstrument(pkgPath string) (isXgo bool, allow bool) {
	_, ok := goinfo.PkgWithinModule(pkgPath, "runtime")
	if ok {
		return false, false
	}
	if isXgoRuntimePkg(pkgPath) {
		return true, false
	}
	if neverInstrumentPkgs[pkgPath] {
		return false, false
	}
	return false, true
}

func isXgoRuntimePkg(pkgPath string) bool {
	// avoid instrument runtime package
	suffix, ok := goinfo.PkgWithinModule(pkgPath, constants.RUNTIME_MODULE)
	if !ok {
		return false
	}
	// check if is runtime/test
	_, isTest := goinfo.PkgWithinModule(suffix, "test")
	return !isTest
}

func GetPkgConfig(pkgPath string) *PkgConfig {
	cfgValue, ok := defaultStdPkgConfig[pkgPath]
	if !ok {
		return nil
	}
	return &cfgValue
}

var neverInstrumentPkgs = map[string]bool{
	"unsafe":      true,
	"runtime":     true,
	"syscall":     true,
	"reflect":     true,
	"sync":        true,
	"sync/atomic": true,
	// testing is not harmful
	// but may cause infinite loop?
	// we may dig later or just add some whitelist
	"testing": true,
}

var defaultStdPkgConfig = map[string]PkgConfig{
	"os": {
		WhitelistFunc: map[string]bool{
			// starts with Get
			"OpenFile":  true,
			"ReadFile":  true,
			"WriteFile": true,
		},
		WhitelistFuncPrefix: []string{"Get"},
	},
	"io": {
		WhitelistFunc: map[string]bool{
			"ReadAll": true,
		},
	},
	"io/ioutil": {
		WhitelistFunc: map[string]bool{
			"ReadAll":  true,
			"ReadFile": true,
			"ReadDir":  true,
		},
	},
	"time": {
		WhitelistFunc: map[string]bool{
			"Now": true,
			// time.Sleep is special:
			//  if trapped like normal functions
			//    runtime/time.go:178:6: ns escapes to heap, not allowed in runtime
			// there are special handling of this, see instrument_runtime/time.go
			"Sleep":       true, // NOTE: time.Sleep links to runtime.timeSleep
			"NewTicker":   true,
			"Time.Format": true,
		},
	},
	"os/exec": {
		WhitelistFunc: map[string]bool{
			"Command":       true,
			"(*Cmd).Run":    true,
			"(*Cmd).Output": true,
			"(*Cmd).Start":  true,
		},
	},
	"net/http": {
		WhitelistFunc: map[string]bool{
			"Get":  true,
			"Head": true,
			"Post": true,
			// Sever
			"Serve":           true,
			"Handle":          true,
			"(*Client).Do":    true,
			"(*Server).Close": true,
		},
	},
	"net": {
		WhitelistFuncPrefix: []string{"(*Dialer).Dial", "Dial"},
	},
}
