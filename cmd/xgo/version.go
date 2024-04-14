package main

import "fmt"

const VERSION = "1.0.24"
const REVISION = "a9cbbd937997b1473a5f70f0131927adcdbf3b79+1"
const NUMBER = 181

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
