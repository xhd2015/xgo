package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "5de8ef30a2f64654a508ac88b5ce8f88dcdb2874+1"
const NUMBER = 139

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
