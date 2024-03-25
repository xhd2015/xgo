package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "b45bd7523aad9c6907afab55e22d5f7ffa730e1b+1"
const NUMBER = 107

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
