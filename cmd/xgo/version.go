package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "f3e4e310320226a1de91a82ea3555dd28c7f306a+1"
const NUMBER = 256

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
