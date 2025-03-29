package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/instrument/load"
	"github.com/xhd2015/xgo/support/instrument/overlay"
)

// usage:
//
//	go run ./support/instrument/load/example --dir x ./...
func main() {
	var args []string
	for i, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			continue
		}
		args = os.Args[i:]
	}
	if len(args) == 0 {
		args = []string{"."}
	}

	n := len(args)
	var dir string
	var remainArgs []string
	for i := 0; i < n; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			remainArgs = append(remainArgs, arg)
			continue
		}
		switch arg {
		case "--dir":
			if i+1 >= n {
				panic("missing dir")
			}
			dir = args[i+1]
			i++
		default:
			panic(fmt.Sprintf("unknown arg: %s", arg))
		}
	}
	// Create an empty overlay using MakeOverlay
	emptyOverlay := overlay.MakeOverlay()

	loadArgs := remainArgs
	if len(loadArgs) == 0 {
		loadArgs = []string{"."}
	}
	// Call LoadPackages with the current directory
	begin := time.Now()
	pkgs, err := load.LoadPackages(loadArgs, load.LoadOptions{
		Dir:     dir,
		Overlay: emptyOverlay,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var numFiles int
	for _, pkg := range pkgs.Packages {
		numFiles += len(pkg.Files)
	}

	// Performance:
	//     Number of packages: 474, Number of files: 3545, cost: 1.1984535s

	// Print basic information about the loaded packages
	fmt.Printf("Number of packages: %d, Number of files: %d, cost: %v\n", len(pkgs.Packages), numFiles, time.Since(begin))
	if len(pkgs.Packages) > 0 {
		fmt.Printf("Package name: %s\n", pkgs.Packages[0].GoPackage.Name)
	}
	// Output:
	// ...
}
