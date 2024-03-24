package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "5dd9a8148b726ed806c30a1644649b13aa60164a+1"
const NUMBER = 100

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
