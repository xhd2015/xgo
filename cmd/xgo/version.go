package main

import "fmt"

const VERSION = "1.0.7"
const REVISION = "d6c2b3d82f39c5ec56856db102f7957cb378853f+1"
const NUMBER = 116

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
