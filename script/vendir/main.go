package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/goinfo"
)

const help = `
vendir helps to create third party vendor dependency without
introducing changes to go.mod.

Usage: 
  vendir create [OPTIONS] <dir> <target_vendor_dir>
  vendir rewrite-file [OPTIONS] <file> <target_vendor_dir>
  vendir rewrite-path <path> <target_vendor_dir>
  vendir help

Arguments:
  <dir> should be either:
  - the root dir containing a go.mod and a vendor dir, or
  - the vendor directory

Options:
  --help   show help message

Example:
  $ go run github.com/xhd2015/xgo/script/vendir create ./x/src ./x/third_party_vendir
`

// usage:
//
//	go run ./script/vendir create ./script/vendir/testdata/src ./script/vendir/testdata/third_party_vendir
func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires cmd")
	}
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "create":
		return createVendor(args)
	case "rewrite-file":
		return rewriteFile(args)
	case "rewrite-path":
		return rewritePath(args)
	case "help":
		fmt.Println(strings.TrimSpace(help))
		return nil
	default:
		return fmt.Errorf("unrecognized cmd: %s, check help", cmd)
	}
}

func createVendor(args []string) error {
	var remainArgs []string
	var verbose bool
	n := len(args)
	for i := 0; i < n; i++ {
		if args[i] == "--rm" {
			if i+1 >= n {
				return fmt.Errorf("%v requires arg", args[i])
			}
			i++
			continue
		}
		if args[i] == "-v" || args[i] == "--verbose" {
			verbose = true
			continue
		}
		if args[i] == "--help" {
			fmt.Println(strings.TrimSpace(help))
			return nil
		}
		if args[i] == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if strings.HasPrefix(args[i], "-") {
			return fmt.Errorf("unrecognized flag: %v", args[i])
		}
		remainArgs = append(remainArgs, args[i])
	}
	if len(remainArgs) < 2 {
		return fmt.Errorf("usage: vendir create <dir> <target_dir>")
	}

	dir := remainArgs[0]
	targetVendorDir := remainArgs[1]

	_, targetStatErr := os.Stat(targetVendorDir)
	if !os.IsNotExist(targetStatErr) {
		return fmt.Errorf("%s already exists, remove it before create", targetVendorDir)
	}
	goMod := filepath.Join(dir, "go.mod")
	vendorDir := filepath.Join(dir, "vendor")
	_, err := os.Stat(goMod)
	if err != nil {
		return err
	}
	_, err = os.Stat(vendorDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(targetVendorDir, 0755)
	if err != nil {
		return err
	}

	modPath, err := goinfo.ResolveModPath(targetVendorDir)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "rewrite prefix: %s\n", modPath)
	}
	rw, err := initRewriter(modPath + "/")
	if err != nil {
		return err
	}

	err = filecopy.Copy(vendorDir, targetVendorDir)
	if err != nil {
		return err
	}
	// traverse all go files, and rewrite
	return filepath.Walk(targetVendorDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		newCode, err := rw.rewriteFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return os.WriteFile(path, []byte(newCode), info.Mode())
	})
}

func rewriteFile(args []string) error {
	var some string

	var remainArgs []string
	n := len(args)
	for i := 0; i < n; i++ {
		if args[i] == "--some" {
			if i+1 >= n {
				return fmt.Errorf("%v requires arg", args[i])
			}
			some = args[i+1]
			i++
			continue
		}
		if args[i] == "--help" {
			fmt.Println(strings.TrimSpace(help))
			return nil
		}
		if args[i] == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if strings.HasPrefix(args[i], "-") {
			return fmt.Errorf("unrecognized flag: %v", args[i])
		}
		remainArgs = append(remainArgs, args[i])
	}
	if len(remainArgs) < 2 {
		return fmt.Errorf("usage: vendir <file> <target_dir>")
	}

	file := remainArgs[0]
	targetDir := remainArgs[1]

	stat, err := os.Stat(file)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is not a file", file)
	}

	modPath, err := goinfo.ResolveModPath(targetDir)
	if err != nil {
		return err
	}
	rw, err := initRewriter(modPath + "/")
	if err != nil {
		return err
	}
	// rewrite file
	code, err := rw.rewriteFile(file)
	if err != nil {
		return err
	}
	fmt.Println(code)

	_ = some

	return nil
}

func rewritePath(args []string) error {
	remainArgs := args
	if len(remainArgs) < 2 {
		return fmt.Errorf("usage: vendir rewrite-path <path> <target_dir>")
	}

	path := remainArgs[0]
	targetDir := remainArgs[1]

	modPath, err := goinfo.ResolveModPath(targetDir)
	if err != nil {
		return err
	}
	rw, err := initRewriter(modPath + "/")
	if err != nil {
		return err
	}

	fmt.Println(rw.rewritePath(path))
	return nil
}
