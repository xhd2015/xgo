package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "e7db17d0725927b95d821d84f5a2935e6e339bcd+1"
const NUMBER = 140

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
