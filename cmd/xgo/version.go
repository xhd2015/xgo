package main

import "fmt"

const VERSION = "1.0.4"
const REVISION = "5834fe9106ac8ef652fa99d1b34597bef941c34a+1"
const NUMBER = 93

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
