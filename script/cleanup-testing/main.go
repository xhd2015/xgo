package main

import (
	"github.com/xhd2015/xgo/script/lib"
)

func main() {
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
	lib.Logf("found %d dirs (%s)", bCount, lib.FormatSize(bSize))

	removed, freedBytes := lib.CleanupXgoTempDirs()
	lib.Logf("removed %d dirs, freed %s", removed, lib.FormatSize(freedBytes))
}
