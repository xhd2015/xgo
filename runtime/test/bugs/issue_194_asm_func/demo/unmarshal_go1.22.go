//go:build !go1.23
// +build !go1.23

// this sonic thing does not support go1.23 yet
package main

import (
	"github.com/bytedance/sonic"
)

const goMajor = 1
const goMinor = 22

func unmarshal(body []byte, u interface{}) {
	if body == nil {
		return
	}
	sonic.Unmarshal(body, u)
}
