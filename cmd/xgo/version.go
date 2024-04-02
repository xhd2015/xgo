package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "9144e01f0f4847402d24c04997c8077fbb3ae85e+1"
const NUMBER = 146

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
