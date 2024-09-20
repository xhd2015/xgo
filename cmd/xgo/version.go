package main

import "fmt"

// auto updated
const VERSION = "1.0.48"
const REVISION = "40aa40fc76231d2c9ee681be5456b26d4255f123+1"
const NUMBER = 307

// manually updated
const CORE_VERSION = "1.0.48"
const CORE_REVISION = "ee7e3078596587e9734a1e6f208d258b8c6fa090+1"
const CORE_NUMBER = 305

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
