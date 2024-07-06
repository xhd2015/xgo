package main

import "fmt"

// auto updated
const VERSION = "1.0.44"
const REVISION = "7ff3bd301202f9d81cfeeb72c34dfd8c46f7d36e+1"
const NUMBER = 286

// manually updated
const CORE_VERSION = "1.0.43"
const CORE_REVISION = "f1cf6698521d5b43da06f012ac3ba5afb1308d27+1"
const CORE_NUMBER = 280

func getRevision() string {
	return formatRevision(VERSION, REVISION, NUMBER)
}

func getCoreRevision() string {
	return formatRevision(CORE_VERSION, CORE_REVISION, CORE_NUMBER)
}

func formatRevision(version string, revision string, number int) string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", version, revision, revSuffix, number)
}
