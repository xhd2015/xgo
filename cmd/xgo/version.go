package main

import "fmt"

const VERSION = "1.0.12"
const REVISION = "f3d7271450fef6b7575368a82d6fe254c894a97e+1"
const NUMBER = 147

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
