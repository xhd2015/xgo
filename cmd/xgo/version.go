package main

import "fmt"

const VERSION = "1.0.2"
const REVISION = "142248b8cb14931e12cebb21385873a8f514b689+1"
const NUMBER = 84

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
