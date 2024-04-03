package main

import "fmt"

const VERSION = "1.0.17"
const REVISION = "77848295e3d73a3eba8ce2bbd1b95d8c988929d5+1"
const NUMBER = 156

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
