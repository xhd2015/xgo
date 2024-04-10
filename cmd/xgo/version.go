package main

import "fmt"

const VERSION = "1.0.20"
const REVISION = "7ab84d83e8d847f0c2307a44866705581ba5cbbe+1"
const NUMBER = 172

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
