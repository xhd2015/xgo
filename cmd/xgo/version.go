package main

import "fmt"

const VERSION = "1.0.8"
const REVISION = "c19fe042380d7ebb280712f232b699c1321932eb+1"
const NUMBER = 121

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
