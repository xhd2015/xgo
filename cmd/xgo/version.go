package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "7fe40f22001bdb57b813f9eb00772efc84c9f27d+1"
const NUMBER = 110

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
