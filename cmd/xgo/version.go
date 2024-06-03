package main

import "fmt"

const VERSION = "1.0.39"
const REVISION = "4d63b2ed7ce161f17d1c311a9b536e82a01769f0+1"
const NUMBER = 259

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
