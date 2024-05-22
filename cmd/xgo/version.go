package main

import "fmt"

const VERSION = "1.0.36"
const REVISION = "48d5fefe9c2c051c940e088429f9253b80a65305+1"
const NUMBER = 230

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
