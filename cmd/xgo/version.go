package main

import "fmt"

const VERSION = "1.0.39"
const REVISION = "ceb96b726f89731a8f2bda8239af7c1810e44294+1"
const NUMBER = 264

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
