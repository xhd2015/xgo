package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "7905f4781a92702218362406619193975db8e1a5+1"
const NUMBER = 190

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
