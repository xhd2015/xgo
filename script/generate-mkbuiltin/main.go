package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xhd2015/xgo/instrument/instrument_compiler"
	"github.com/xhd2015/xgo/support/goinfo"
)

func main() {
	var origGoroot, outputGoroot, goVersionStr string
	flag.StringVar(&origGoroot, "orig-goroot", "", "original unmodified GOROOT (used to run go binary)")
	flag.StringVar(&outputGoroot, "output-goroot", "", "instrumented GOROOT (files are written here)")
	flag.StringVar(&goVersionStr, "go-version", "", "go version string, e.g. 'go1.24.7'")
	flag.Parse()

	if origGoroot == "" || outputGoroot == "" || goVersionStr == "" {
		fmt.Fprintf(os.Stderr, "usage: generate-mkbuiltin --orig-goroot=<path> --output-goroot=<path> --go-version=go1.24.7\n")
		os.Exit(1)
	}

	goVersion, err := goinfo.ParseGoVersion("go version " + goVersionStr + " darwin/amd64")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = instrument_compiler.MkBuiltin(origGoroot, outputGoroot, goVersion, instrument_compiler.RuntimeExtraDef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
