package main

import "fmt"

const VERSION = "1.0.0"
const REVISION = "37d37c71b68249a79d6cb3347d41945295ecde57+1"
const NUMBER = 79

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
