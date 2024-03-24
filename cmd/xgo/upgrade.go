package main

import (
	"github.com/xhd2015/xgo/cmd/xgo/upgrade"
)

func handleUpgrade(args []string) error {
	var installDir string
	nArg := len(args)
	for i := 0; i < nArg; i++ {
		arg := args[i]
		if arg == "--install-dir" {
			installDir = args[i+1]
			i++
			continue
		}
	}
	return upgrade.Upgrade(installDir)
}
