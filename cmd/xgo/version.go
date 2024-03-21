package main

import "fmt"

const VERSION = "1.0.4"
const REVISION = "1547d633339c6ac264555d484c051e9ed4582b71+1"
const NUMBER = 89

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
