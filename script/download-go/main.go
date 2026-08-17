package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/xgo/support/downloadgo"
)

// usage: list, download go1.20.0

const goReleaseDir = "go-release"

const helpText = `Usage: download-go <command> [options]

Commands:
  list                    List all available Go versions for download
  download <version>      Download a specific Go version
  go<version>             Shorthand for download go<version>

Options:
  --dir <dir>             Target directory for download (default: go-release)
  -h, --help              Show this help

Examples:
  download-go list
  download-go go1.22.1
  download-go download 1.22.1
  download-go download 1.22.1 --dir ./go-sdks
`

func main() {
	args := os.Args[1:]
	if err := run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(helpText)
		return nil
	}

	var cmd string
	var version string
	var targetDir string
	if len(args) > 0 {
		cmd = args[0]
		if cmd == "download" {
			if len(args) > 1 && (args[1] == "-h" || args[1] == "--help") {
				fmt.Print(helpText)
				return nil
			}
			var err error
			version, targetDir, err = parseDownloadArgs(args[1:])
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(cmd, "go") {
			version = cmd
			cmd = "download"
		}
	}
	if cmd == "" {
		return fmt.Errorf("requires cmd")
	}

	ctx := context.Background()
	if cmd == "list" {
		goVersions, err := downloadgo.List(ctx, downloadgo.ListOptions{})
		if err != nil {
			return err
		}
		for _, goVersion := range goVersions {
			fmt.Printf("go%s\n", goVersion)
		}
		return nil
	}
	if cmd != "download" {
		return fmt.Errorf("unrecognized cmd: %s", cmd)
	}
	if targetDir == "" {
		targetDir = goReleaseDir
	}

	_, err := downloadgo.Download(ctx, version, downloadgo.Options{
		Dir:    targetDir,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	return err
}

func parseDownloadArgs(args []string) (version string, targetDir string, err error) {
	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--dir" {
			if i+1 >= n {
				err = fmt.Errorf("%s requires arg", arg)
				return
			}
			targetDir = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			err = fmt.Errorf("unrecognized flag: %s", arg)
			return
		}
		version = arg
	}
	return
}
