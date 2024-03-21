package main

import "fmt"

const VERSION = "1.0.3"
const REVISION = "1385a83397168d43773577b475bbae38ea44b813+1"
const NUMBER = 87

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
