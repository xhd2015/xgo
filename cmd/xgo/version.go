package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "cbc3deeffd52cddb0f9a845c0e079a658f8c8325+1"
const NUMBER = 145

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
