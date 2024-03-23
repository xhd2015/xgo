package main

import "fmt"

const VERSION = "1.0.5"
const REVISION = "1c2f189bbbc77804d3acab01865e6ef6064f7705+1"
const NUMBER = 95

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
