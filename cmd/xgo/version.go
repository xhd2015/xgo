package main

import "fmt"

const VERSION = "1.0.5"
const REVISION = "a651abf3abfec58ea52a4d774a1b04c6f7c70c1d+1"
const NUMBER = 97

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
