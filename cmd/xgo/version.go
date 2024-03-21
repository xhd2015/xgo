package main

import "fmt"

const VERSION = "1.0.4"
const REVISION = "185ff6be71bab0c6908105127789c7f108ee2437+1"
const NUMBER = 88

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
