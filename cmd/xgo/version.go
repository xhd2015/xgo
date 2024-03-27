package main

import "fmt"

const VERSION = "1.0.7"
const REVISION = "8844707a09ae1544c2ad9db4f83ae7ee34601136+1"
const NUMBER = 113

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
