package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// VERSION is manually updated when needed a new tag
// see also runtime/core/version.go
const VERSION = "1.0.52"
const REVISION = "0e1111a8a3f33ff426c1c5060ed1944ab7279fc3+1"
const NUMBER = 330

// the matching runtime/core's version
// manually updated
const CORE_VERSION = "1.0.52"
const CORE_REVISION = "a6f0088f2e43fe837c905792459dfca4e1022a0b+1"
const CORE_NUMBER = 324

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
