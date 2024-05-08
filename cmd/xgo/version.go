package main

import "fmt"

const VERSION = "1.0.30"
const REVISION = "83fdf348a92806bbdc0b6746fc7d55ae1671dfab+1"
const NUMBER = 209

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
