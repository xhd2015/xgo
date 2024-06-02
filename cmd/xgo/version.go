package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "c77350703e0305824f28a2ff5879e1499e47fc39+1"
const NUMBER = 252

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
