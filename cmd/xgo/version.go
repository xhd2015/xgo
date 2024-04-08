package main

import "fmt"

const VERSION = "1.0.19"
const REVISION = "9f58dbf7230665a152fdc35eca356fb0f269a85d_DEV_2024-04-08T07:03:58Z"
const NUMBER = 170

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
