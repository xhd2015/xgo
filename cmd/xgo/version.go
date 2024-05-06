package main

import "fmt"

const VERSION = "1.0.29"
const REVISION = "5d121a999ca8ce5c744a110d4f42b31ef4490e8f+1"
const NUMBER = 204

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
