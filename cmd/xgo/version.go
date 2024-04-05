package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "f0cc411faaf6a9b12a4140f32e02b5f3c461a91d+1"
const NUMBER = 161

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
