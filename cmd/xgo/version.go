package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "6d745e9f6947feaad746282a0b03a87fc335678b+1"
const NUMBER = 142

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
