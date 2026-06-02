package patch

import (
	"cmd/compile/internal/ir"
	"fmt"
	"os"
	"path/filepath"

	"cmd/compile/internal/xgo_rewrite_internal/patch/funcs"
	"cmd/compile/internal/xgo_rewrite_internal/patch/link"
)

func Patch() {
	appendLog("PATCH called")
	linkFuncs()
	appendLog("PATCH done")
}

func linkFuncs() {
	count := 0
	funcs.ForEach(func(fn *ir.Func) bool {
		link.LinkXgoInit(fn)
		count++
		return true
	})
	appendLog("PATCH linkFuncs iterated %d funcs", count)
}

func appendLog(format string, args ...interface{}) {
	logPath := filepath.Join(os.TempDir(), "xgo_link_init_debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// try fallback: write to stderr on open failure
		fmt.Fprintf(os.Stderr, "XGO_LOG_ERROR: cannot open %s: %v\n", logPath, err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, format+"\n", args...)
}
