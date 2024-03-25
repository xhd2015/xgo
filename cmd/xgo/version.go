package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "42bc8080ae3bca80a1f32b5d60bead4a55ea942c+1"
const NUMBER = 106

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
