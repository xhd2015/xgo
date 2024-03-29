package main

import "fmt"

const VERSION = "1.0.8"
const REVISION = "0f3164c52f7cf42a68918206f22ffb36643c6d98+1"
const NUMBER = 120

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
