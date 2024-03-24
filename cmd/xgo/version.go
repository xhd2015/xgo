package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "66b6b96f6da45b9d81b2392d8f88077f3d728f77+1"
const NUMBER = 101

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
