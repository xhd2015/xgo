package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "9028d0e560172d8b5c5ccba73491eed2886ddd6c+1"
const NUMBER = 189

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
