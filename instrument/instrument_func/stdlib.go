package instrument_func

import (
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/support/goinfo"
)

type stdPkgConfig struct {
	neverInstrument     bool
	whitelistFunc       map[string]bool
	whitelistFuncPrefix []string
}

func IsPkgAllowed(pkg *edit.Package) bool {
	pkgPath := pkg.LoadPackage.GoPackage.ImportPath
	cfg, ok := stdPkgConfigMapping[pkgPath]
	if ok && cfg.neverInstrument {
		return false
	}
	if pkg.LoadPackage.GoPackage.Standard {
		// stdlib whitelist mode
		return ok
	}
	// avoid instrument runtime package
	suffix, isXgoRuntime := goinfo.PkgWithinModule(pkgPath, constants.RUNTIME_MODULE)
	if isXgoRuntime {
		// check if is runtime/test
		_, ok := goinfo.PkgWithinModule(suffix, "test")
		if ok {
			return true
		}
		// a regular runtime package
		return false
	}
	return true
}

func IsPkgNeverInstrument(pkgPath string) bool {
	cfg, ok := stdPkgConfigMapping[pkgPath]
	if ok && cfg.neverInstrument {
		return true
	}
	return false
}

var stdPkgConfigMapping = map[string]stdPkgConfig{
	"os": {
		whitelistFunc: map[string]bool{
			// starts with Get
			"OpenFile":  true,
			"ReadFile":  true,
			"WriteFile": true,
		},
		whitelistFuncPrefix: []string{"Get"},
	},
	"io": {
		whitelistFunc: map[string]bool{
			"ReadAll": true,
		},
	},
	"io/ioutil": {
		whitelistFunc: map[string]bool{
			"ReadAll":  true,
			"ReadFile": true,
			"ReadDir":  true,
		},
	},
	"time": {
		whitelistFunc: map[string]bool{
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
		whitelistFunc: map[string]bool{
			"Command":       true,
			"(*Cmd).Run":    true,
			"(*Cmd).Output": true,
			"(*Cmd).Start":  true,
		},
	},
	"net/http": {
		whitelistFunc: map[string]bool{
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
		whitelistFuncPrefix: []string{"(*Dialer).Dial", "Dial"},
	},
	"syscall": {
		neverInstrument: true,
	},
	"reflect": {
		neverInstrument: true,
	},
	"sync": {
		neverInstrument: true,
	},
	"sync/atomic": {
		neverInstrument: true,
	},

	// testing is not harmful
	// but may cause infinite loop?
	// we may dig later or just add some whitelist
	"testing": {
		neverInstrument: true,
	},
	"unsafe": {
		neverInstrument: true,
	},
	"runtime": {
		neverInstrument: true,
	},
}
