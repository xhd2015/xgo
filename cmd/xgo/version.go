package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "d590df444a20fcbe9a1a666bec64856b3211e72f+1"
const NUMBER = 159

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
