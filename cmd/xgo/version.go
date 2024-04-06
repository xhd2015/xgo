package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "1211c519c8005ddbd66189cf64e958aa69e5789f+1"
const NUMBER = 163

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
