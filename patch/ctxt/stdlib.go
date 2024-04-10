package ctxt

import "strings"

var stdWhitelist = map[string]map[string]bool{
	// "runtime": map[string]bool{
	// "timeSleep": true,
	// },
	"os": map[string]bool{
		// starts with Get
		"OpenFile":  true,
		"ReadFile":  true,
		"WriteFile": true,
	},
	"io": map[string]bool{
		"ReadAll": true,
	},
	"io/ioutil": map[string]bool{
		"ReadAll":  true,
		"ReadFile": true,
		"ReadDir":  true,
	},
	"time": map[string]bool{
		"Now": true,
		// time.Sleep is special:
		//  if trapped like normal functions
		//    runtime/time.go:178:6: ns escapes to heap, not allowed in runtime
		// there are special handling of this, see cmd/xgo/patch_runtime patchRuntimeTime
		"Sleep":       true, // NOTE: time.Sleep links to runtime.timeSleep
		"NewTicker":   true,
		"Time.Format": true,
	},
	"os/exec": map[string]bool{
		"Command":       true,
		"(*Cmd).Run":    true,
		"(*Cmd).Output": true,
		"(*Cmd).Start":  true,
	},
	"net/http": map[string]bool{
		"Get":  true,
		"Head": true,
		"Post": true,
		// Sever
		"Serve":           true,
		"Handle":          true,
		"(*Client).Do":    true,
		"(*Server).Close": true,
	},
	"net": map[string]bool{
		// starts with Dial
	},
	"encoding/json": map[string]bool{
		"newTypeEncoder": true,
	},
}

func allowStdFunc(pkgPath string, funcName string) bool {
	if stdWhitelist[pkgPath][funcName] {
		return true
	}
	switch pkgPath {
	case "os":
		return strings.HasPrefix(funcName, "Get")
	case "net":
		return strings.HasPrefix(funcName, "Dial")
	}
	// by default block all
	return false
}
