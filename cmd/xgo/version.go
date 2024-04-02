package main

import "fmt"

const VERSION = "1.0.12"
const REVISION = "3f5b60b1befac59464d3efbe2e9c2b5eeb35cf7b+1"
const NUMBER = 150

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
