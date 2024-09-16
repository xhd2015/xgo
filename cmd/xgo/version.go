package main

import "fmt"

// auto updated
const VERSION = "1.0.48"
const REVISION = "e82e0eba1db860999c91f1e94af190e17e1c9ce5+1"
const NUMBER = 306

// manually updated
const CORE_VERSION = "1.0.48"
const CORE_REVISION = "ee7e3078596587e9734a1e6f208d258b8c6fa090+1"
const CORE_NUMBER = 305

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
