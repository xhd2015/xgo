package main

import "fmt"

// REVISION and NUMBER are auto updated when run 'git commit'
// they correspond to a unique commit.
// VERSION is manually updated when needed a new tag
// if you did not install git hooks, you can manually update them
const VERSION = "1.1.3"
const REVISION = "06607072627f353832c23bf60b0554f348859192+1"
const NUMBER = 424

// Rationale: xgo consists of these modules:
//
// - cmd/xgo
// - cmd/xgo/test-explorer
// - cmd/xgo/trace
// - runtime
// - cmd/xgo/runtime_gen (generated from runtime)
//
// They share the same version, but not every commit
// affects the instrumentation.
//
// To distinguish them,
// we use CORE_REVISION and CORE_NUMBER.
// The core version denotes the core functionality
// of:
//   - cmd/xgo
//   - runtime
//
// general rule is, if you changed anything that
// affects cmd/xgo or runtime, bump the CORE_REVISION
// and CORE_NUMBER.
//
// The CORE_VERSION is manually updated
// once updated, it will be copied to runtime/core/version.go
//
// steps to update:
//  1. run `go run ./script/generate cmd/xgo/version.go`
//  2. copy from REVISION and NUMBER above to the following constants
//  3. run `go run ./script/generate runtime/core/version.go`
//
// finally you will find that the two groups of constants are the same.
const CORE_VERSION = "1.1.2"
const CORE_REVISION = "300cc4e662fd32e51c0c17af3319335d41dac57b+1"
const CORE_NUMBER = 422

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
