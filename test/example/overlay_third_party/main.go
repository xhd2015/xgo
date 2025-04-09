package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/git"
	"github.com/xhd2015/xgo/support/goinfo"
	"golang.org/x/example/hello/reverse"
)

func greet(name string) string {
	return "hello " + name
}

// go run ./ update --> generate overlay.json
func main() {
	if len(os.Args) > 1 && os.Args[1] == "update" {
		rootDir, err := git.ShowTopLevel("")
		if err != nil {
			panic(err)
		}
		dir := filepath.Join(rootDir, "test", "example", "overlay_third_party")
		const PKG = "golang.org/x/example/hello/reverse"
		pkgs, err := goinfo.ListPackages([]string{PKG}, goinfo.LoadPackageOptions{
			Dir: dir,
		})
		if err != nil {
			panic(err)
		}
		var pkgDir string
		for _, pkg := range pkgs {
			if pkg.ImportPath != PKG {
				continue
			}
			pkgDir = pkg.Dir
			break
		}
		if pkgDir == "" {
			panic("package " + PKG + " not found")
		}
		reverseFile := filepath.Join(pkgDir, "reverse.go")

		overlay := map[string]interface{}{
			"Replace": map[string]string{
				reverseFile: filepath.Join(dir, "overlay", "reverse.go"),
			},
		}
		overlayBytes, err := json.Marshal(overlay)
		if err != nil {
			panic(err)
		}
		os.WriteFile(filepath.Join(dir, "overlay.json"), overlayBytes, 0644)
		fmt.Fprintf(os.Stderr, "pkgDir: %s\n", pkgDir)
		return
	}
	fmt.Println(reverse.String(greet("world")))
}
