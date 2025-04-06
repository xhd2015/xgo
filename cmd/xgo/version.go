package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// they correspond to a unique commit.
// VERSION is manually updated when needed a new tag
// if you did not install git hooks, you can manually update them
const VERSION = "1.1.0"
const REVISION = "e6145d35e5bb55867b84e35fc55d5d66e7e2f734+1"
const NUMBER = 375

// the corresponding runtime/core's version
// manually updated
// once updated, they will be copied to runtime/core/version.go
//
// general guidelines is:
//
//	when there is some runtime update, bump the following
//	version to be same with above.
//
// run `go run ./script/generate runtime/core/version.go` will
// do this job.
// usually do a `git commit` first, then
// `go run ./script/generate --amend`, then
// `git commit --amend --no-edit` everything will keep in sync.
const CORE_VERSION = "1.1.0"
const CORE_REVISION = "e6145d35e5bb55867b84e35fc55d5d66e7e2f734+1"
const CORE_NUMBER = 375

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
