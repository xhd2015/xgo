package main

import "fmt"

const VERSION = "1.0.24"
const REVISION = "b70bf6dce3af4317cb6a3b2b18dc30e2b36b8afa+1"
const NUMBER = 179

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
