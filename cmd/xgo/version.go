package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// VERSION is manully updated when needed a new tag
// see also runtime/core/version.go
const VERSION = "1.0.49"
const REVISION = "bfe5f14d7bfaf52b5dfecb95aaaa717e0ff4d3c9+1"
const NUMBER = 310

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
