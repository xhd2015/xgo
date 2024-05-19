package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "d3b65d5d532f17ee01275785d7872692f3a7b76a+1"
const NUMBER = 221

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
