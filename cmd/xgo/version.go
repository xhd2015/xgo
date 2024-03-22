package main

import "fmt"

const VERSION = "1.0.4"
const REVISION = "02a21589a89dd64adafd02d34fd321d6106671e4+1"
const NUMBER = 90

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
