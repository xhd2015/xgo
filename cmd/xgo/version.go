package main

import "fmt"

const VERSION = "1.0.17"
const REVISION = "d56f35be1d57233e6b58145879ebe0fedf976361+1"
const NUMBER = 155

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
