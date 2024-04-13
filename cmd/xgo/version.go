package main

import "fmt"

const VERSION = "1.0.23"
const REVISION = "e8daf6b3b430f0be7a30aa02bf0f8334c93537ec+1"
const NUMBER = 178

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
