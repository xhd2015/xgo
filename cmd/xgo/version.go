package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// they correspond to a unique commit.
// VERSION is manually updated when needed a new tag
// if you did not install git hooks, you can manually update them
const VERSION = "1.1.1"
const REVISION = "b48bc9f28e1ecad8f91be9cb7c536a49c6fce786+1"
const NUMBER = 390

// TODO: decouple CORE_VERSION here and that in runtime/core/version.go
// because this now only indicates lowest working version required by xgo.
//
// the CORE_VERSION marks the lowest working version required by xgo.
// even if CORE_VERSION is lower than xgo's VERSION, xgo can still
// work with the newest xgo/runtime.
// As long as this holds, we don't need to change CORE_VERSION.
//
// Rationale: there is no reason to force user to upgrade xgo,
// which is only upgraded for new feature. if that feature
// is not required by user, user don't have to upgrade xgo.
//
// And xgo/runtime's API is quite stable except the internal
// package, which we can rewrite in newer xgo.
//
// So in conclusion, across the life of a major version,
// CORE_VERSION can remain the same.
//
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
const CORE_REVISION = "f5351b88ed22ac3d20f85e33b1fdb3fe624ce61c+1"
const CORE_NUMBER = 388

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
