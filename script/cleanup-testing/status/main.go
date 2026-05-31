package main

import (
	"github.com/xhd2015/xgo/script/lib"
)

func main() {
	entries := lib.ListXgoTempDirs()
	if len(entries) == 0 {
		lib.Logf("no xgo temp dirs found")
		return
	}

	lib.Logf("detailed listing:")
	for _, e := range entries {
		lib.Logf("  %s  %s", lib.FormatSize(e.Size), e.Name)
	}

	lib.Logf("")
	lib.Logf("summary by pattern:")
	stats := lib.StatsByPattern()
	var total int64
	for _, s := range stats {
		lib.Logf("  %-25s  %3d dirs  %s", s.Pattern, s.Count, lib.FormatSize(s.Size))
		total += s.Size
	}
	lib.Logf("  %-25s  %3d files  %s", "TOTAL", len(entries), lib.FormatSize(total))
}
