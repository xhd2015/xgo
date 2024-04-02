package main

import "fmt"

const VERSION = "1.0.14"
const REVISION = "649cbfce18b726d457516babe9f10a82e07c023d+1"
const NUMBER = 149

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
