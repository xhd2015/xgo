package main

import "fmt"

const VERSION = "1.0.19"
const REVISION = "75c04e25cd9ccc811d6893fc0c0c02df889cad66+1"
const NUMBER = 171

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
