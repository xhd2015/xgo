package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "ef6e707d98b8da66eba11afff2039c69f611971f+1"
const NUMBER = 109

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
