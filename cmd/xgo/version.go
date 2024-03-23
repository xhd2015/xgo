package main

import "fmt"

const VERSION = "1.0.5"
const REVISION = "6812c9db2c530a24f66d5304a23f933f59b2edae+1"
const NUMBER = 94

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
