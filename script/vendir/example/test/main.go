package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/script/vendir/example/internal/third_party_vendir/github.com/xhd2015/less-gen/go/load"
	"github.com/xhd2015/xgo/script/vendir/example/internal/third_party_vendir/github.com/xhd2015/less-gen/go/project"
)

func main() {
	project, err := project.Load([]string{"./test"}, &load.LoadOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	pkg, err := project.GetOnlyEntryPackage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println(pkg.GoPkg().PkgPath)
}
