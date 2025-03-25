package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// VERSION is manually updated when needed a new tag
// if you did not install git hooks, you can manually update them
const VERSION = "1.0.53"
const REVISION = "8fb7775ce713b6fda1d78163735055d521afc33e+1"
const NUMBER = 334

// the wanted runtime/core's version
// manually updated
// see runtime/core/version.go
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
