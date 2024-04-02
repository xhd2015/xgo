package main

import "fmt"

const VERSION = "1.0.13"
const REVISION = "662c91235bb493a3eaa604f06889921257074939+1"
const NUMBER = 148

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
