package main

import "fmt"

const VERSION = "1.0.1"
const REVISION = "4a87957ab7175717205880ca95086e259f342c25+1"
const NUMBER = 80

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
