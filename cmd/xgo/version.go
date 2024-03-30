package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "fab09bf4d2639a63f487304bd91d4eaa40788dbb+1"
const NUMBER = 141

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
