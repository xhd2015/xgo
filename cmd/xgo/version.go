package main

import "fmt"

const VERSION = "1.0.7"
const REVISION = "6101b0a1a4b369d427afd75fed01273d0e09b933+1"
const NUMBER = 114

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
