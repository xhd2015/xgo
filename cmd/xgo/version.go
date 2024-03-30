package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "d5a5e38d579cd502f5cde96570afad9e7d1c21f6+1"
const NUMBER = 138

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
