package main

import "fmt"

const VERSION = "1.0.29"
const REVISION = "818c18ff3d125cf9734fe4d554655c922d645534+1"
const NUMBER = 205

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
