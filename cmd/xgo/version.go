package main

import "fmt"

// auto updated
const VERSION = "1.0.42"
const REVISION = "e185363a861e52ed68adc1dd2f029b530732de51+1"
const NUMBER = 275

// manually updated
const CORE_VERSION = "1.0.42"
const CORE_REVISION = "e185363a861e52ed68adc1dd2f029b530732de51+1"
const CORE_NUMBER = 275

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
