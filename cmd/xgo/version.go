package main

import "fmt"

const VERSION = "1.0.4"
const REVISION = "1685f0ddb83838bc9e0d3c9c4983f5ae7d25e1cd+1"
const NUMBER = 92

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
