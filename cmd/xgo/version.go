package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "d7c7e42a5d12600ec45874e6bace2f943cd915e4+1"
const NUMBER = 102

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
