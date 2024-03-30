package main

import "fmt"

const VERSION = "1.0.10"
const REVISION = "32788b163e5dd0f4ab1c34ebd5ff1781237ccdd1+1"
const NUMBER = 124

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
