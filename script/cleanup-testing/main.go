package main

import (
	"github.com/xhd2015/xgo/script/lib"
)

func main() {
	lib.Logf("cleaning git-versioned GOROOTs...")
	gitResults := lib.CleanGitGoroots()
	for _, r := range gitResults {
		if r.Err != nil {
			lib.Logf("ERROR: %s: %v", r.Dir, r.Err)
		} else {
			lib.Logf("OK: %s", r.Dir)
		}
	}

	before := lib.ListXgoTempDirs()
	if len(before) == 0 {
		lib.Logf("no xgo temp dirs found")
		return
	}

	bCount := len(before)
	var bSize int64
	for _, e := range before {
		bSize += e.Size
	}
	lib.Logf("found %d temp dirs (%s)", bCount, lib.FormatSize(bSize))

	removed, _ := lib.CleanupXgoTempDirs()
	lib.Logf("removed %d temp dirs", removed)
}
