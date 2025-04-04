package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// they correspond to a unique commit.
// VERSION is manually updated when needed a new tag
// if you did not install git hooks, you can manually update them
const VERSION = "1.1.0"
const REVISION = "792e4144baa304be7de04eb0b036a558a7b7b168+1"
const NUMBER = 365

// the wanted runtime/core's version
// manually updated
// see runtime/core/version.go
const CORE_VERSION = "1.1.0"
const CORE_REVISION = "c4a8899c3c25a6701effd940ac2950ad65aed6ab+1"
const CORE_NUMBER = 327

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
