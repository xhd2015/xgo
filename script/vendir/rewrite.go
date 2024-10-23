package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/edit/goedit"
)

type rewriter struct {
	stdPkgMapping map[string]bool
	prefix        string
}

func initRewriter(prefix string) (*rewriter, error) {
	if prefix == "" {
		return nil, fmt.Errorf("requires prefix")
	}
	goroot, err := getGoroot()
	if err != nil {
		return nil, err
	}
	stdPkgs, err := listStdPkgs(goroot)
	if err != nil {
		return nil, err
	}
	stdPkgMapping := make(map[string]bool, len(stdPkgs))
	for _, pkg := range stdPkgs {
		stdPkgMapping[pkg] = true
	}

	return &rewriter{
		stdPkgMapping: stdPkgMapping,
		prefix:        prefix,
	}, nil
}

func (c *rewriter) rewritePath(path string) string {
	if path == "" {
		return ""
	}
	// relative import
	switch path[0] {
	case '/', '.':
		return path
	}
	if c.stdPkgMapping[path] {
		return path
	}
	return c.prefix + path
}

func (c *rewriter) rewriteFile(file string) (string, error) {
	code, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return c.rewriteCode(string(code))
}

func (c *rewriter) rewriteCode(code string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.ImportsOnly)
	if err != nil {
		return "", err
	}

	edit := goedit.New(fset, code)
	for _, imp := range file.Imports {
		pkg, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return "", err
		}
		newPkg := c.rewritePath(pkg)
		if newPkg != pkg {
			edit.Replace(imp.Path.Pos(), imp.Path.End(), strconv.Quote(newPkg))
		}
	}
	return edit.String(), nil
}

func getGoroot() (string, error) {
	goroot, err := cmd.Output("go", "env", "GOROOT")
	if err != nil {
		return "", err
	}
	goroot = strings.TrimSpace(goroot)
	if goroot == "" {
		return "", fmt.Errorf("cannot get 'go env GOROOT'")
	}
	return goroot, nil
}

func listStdPkgs(goroot string) ([]string, error) {
	if goroot == "" {
		return nil, fmt.Errorf("requires GOROOT")
	}
	res, err := cmd.Dir(filepath.Join(goroot, "src")).Output("go", "list", "./...")
	if err != nil {
		return nil, err
	}
	list := strings.Split(res, "\n")
	j := 0
	for i := 0; i < len(list); i++ {
		e := strings.TrimSpace(list[i])
		if e != "" {
			list[j] = e
			j++
		}
	}
	return list[:j], nil
}
