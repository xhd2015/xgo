package main

import "fmt"

const VERSION = "1.0.7"
const REVISION = "52fa68beb627b6fc4246eeeeab191e2bc562cb8a+1"
const NUMBER = 115

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
