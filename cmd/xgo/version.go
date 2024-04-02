package main

import "fmt"

const VERSION = "1.0.15"
const REVISION = "ee229bd12d0d0fdb24f26602651b01dbd76c7cf1+1"
const NUMBER = 150

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
