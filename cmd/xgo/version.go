package main

import "fmt"

const VERSION = "1.0.3"
const REVISION = "c96428ea36bd88f938dbbe5e2b0213ad68c1431b+1"
const NUMBER = 86

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
