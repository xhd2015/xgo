package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "d7fa92e870f1be98c43f42107ad24f4a3152ef5a+1"
const NUMBER = 157

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
