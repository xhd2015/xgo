package main

import "fmt"

const VERSION = "1.0.39"
const REVISION = "e6ba142b8a4c344bfd9e009090d1666635e1c9e5+1"
const NUMBER = 261

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
