package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// they correspond to a unique commit.
// VERSION is manually updated when needed a new tag
// if you did not install git hooks, you can manually update them
const VERSION = "1.1.0"
const REVISION = "6d375e4ed413d62c69af25039a08fb475107aaa1+1"
const NUMBER = 381

// the corresponding runtime/core's version
// manually updated
// once updated, they will be copied to runtime/core/version.go
//
// general guidelines is:
//
//	when there is some runtime update, bump the following
//	version to be same with above.
//
// steps to update:
//  1. run `go run ./script/generate cmd/xgo/version.go`
//  2. copy from REVISION and NUMBER above to the following constants
//  3. run `go run ./script/generate runtime/core/version.go`
//
// finally you will find that the two groups of constants are the same.
const CORE_VERSION = "1.1.0"
const CORE_REVISION = "2d734d2e2c0d29c25babfac2ee6b6473f9f068bf+1"
const CORE_NUMBER = 377

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
