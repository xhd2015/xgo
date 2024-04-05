package main

import (
	"os"

	"github.com/xhd2015/xgo/cmd/xgo/exec_tool"
)

// go1.21.7/pkg/tool/darwin_amd64/compile -o /var/.../_pkg_.a -trimpath /var/...=> -p fmt -std -complete -buildid b_xx -goversion go1.21.7 -c=4 -nolocalimports -importcfg /var/.../importcfg -pack src/A.go src/B.go
// go1.21.7/pkg/tool/darwin_amd64/link -V=full
func main() {
	// os.Arg[0] = exec_tool
	// os.Arg[1] = compile or others...
	args := os.Args[1:]
	exec_tool.Main(args)
}
