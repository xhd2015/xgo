package main

import "fmt"

const VERSION = "1.0.2"
const REVISION = "1030fb3582b5855454eb4c752f8d4c0aec7db6ee+1"
const NUMBER = 83

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
