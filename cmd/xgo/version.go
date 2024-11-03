package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// VERSION is manually updated when needed a new tag
// see also runtime/core/version.go
const VERSION = "1.0.51"
const REVISION = "f43d1801162be8396dd5b73a27b529253a121a88+1"
const NUMBER = 319

// the matching runtime/core's version
// manually updated
const CORE_VERSION = "1.0.49"
const CORE_REVISION = "37977b002ee8cc375e071b7ac23e8bb67a2de64d+1"
const CORE_NUMBER = 308

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
