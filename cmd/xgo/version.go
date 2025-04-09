package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// they correspond to a unique commit.
// VERSION is manually updated when needed a new tag
// if you did not install git hooks, you can manually update them
const VERSION = "1.1.1"
const REVISION = "b385ec2f25f698ce2ee86ec091bf3e4499eb7762+1"
const NUMBER = 383

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
const CORE_VERSION = "1.1.1"
const CORE_REVISION = "b385ec2f25f698ce2ee86ec091bf3e4499eb7762+1"
const CORE_NUMBER = 383

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
