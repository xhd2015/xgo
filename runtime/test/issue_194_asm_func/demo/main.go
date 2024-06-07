package main

import (
	"github.com/bytedance/sonic"
)

func main() {
	unmarshal(nil, nil)
}

func unmarshal(body []byte, u interface{}) {
	if body == nil {
		return
	}
	sonic.Unmarshal(body, u)
}
