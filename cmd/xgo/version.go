package main

import "fmt"

const VERSION = "1.0.16"
const REVISION = "b27192a64f46ddf3ddd2b02ca6fe52c8c4e03ffd+1"
const NUMBER = 154

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
