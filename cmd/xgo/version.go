package main

import "fmt"

const VERSION = "1.0.15"
const REVISION = "05e21215595deccfc08d49cb2d502e0d48b3cf4b+1"
const NUMBER = 151

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
