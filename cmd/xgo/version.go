package main

import "fmt"

const VERSION = "1.0.9"
const REVISION = "99fc85f0589192aa192d793a0b93c82935609fb8+1"
const NUMBER = 122

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
