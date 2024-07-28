package main

import "fmt"

// auto updated
const VERSION = "1.0.47"
const REVISION = "324b023bc90dfb3c23fa9781cbb8f78e2830912b+1"
const NUMBER = 303

// manually updated
const CORE_VERSION = "1.0.47"
const CORE_REVISION = "2b5b421d4069bcde6b70e76b157e98e39634672c+1"
const CORE_NUMBER = 301

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
