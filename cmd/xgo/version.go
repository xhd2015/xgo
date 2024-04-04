package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "e0fff32b787ec10d8c27b655f546a8a43e090b61+1"
const NUMBER = 158

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
