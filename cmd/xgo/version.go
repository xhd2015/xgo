package main

import "fmt"

const VERSION = "1.0.39"
const REVISION = "55af407f826d020bcd28c51663d0879f8d1ca76e+1"
const NUMBER = 258

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
