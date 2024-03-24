package main

import "fmt"

const VERSION = "1.0.5"
const REVISION = "0fa3e3ae52da9151f5b33b141a3f05dde73a8c51+1"
const NUMBER = 96

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
