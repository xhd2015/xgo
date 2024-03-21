package main

import "fmt"

const VERSION = "1.0.2"
const REVISION = "b7119d6422a89527402c0368d7dbc8e2b50741bb+1"
const NUMBER = 85

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
