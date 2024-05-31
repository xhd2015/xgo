package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "d939407b345d1697a3a32cf48c5b659d017c4cff+1"
const NUMBER = 248

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
