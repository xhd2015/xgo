package main

import "fmt"

const VERSION = "1.0.36"
const REVISION = "fce3dfc3c3587abde57944d6b16030437a9d8ce9+1"
const NUMBER = 228

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
