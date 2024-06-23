package main

import "fmt"

// auto updated
const VERSION = "1.0.41"
const REVISION = "e1e4daf465c8474538b38094b2cbfc47fea802bc+1"
const NUMBER = 272

// manually updated
const CORE_VERSION = "1.0.41"
const CORE_REVISION = "35cb77e2af63562938bdd34f94bda831a62d5518+1"
const CORE_NUMBER = 271

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
