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
		err = GenerateOverlay(dir, filepath.Join(dir, "overlay.json"), PKG, "reverse.go", filepath.Join(dir, "overlay", "reverse.go"))
		if err != nil {
			panic(err)
		}
		return
	}
	fmt.Println(reverse.String(greet("world")))
}

func GenerateOverlay(dir string, overlayFile string, pkg string, file string, replacedFile string) error {
	pkgs, err := goinfo.ListPackages([]string{pkg}, goinfo.LoadPackageOptions{
		Dir: dir,
	})
	if err != nil {
		return err
	}
	var pkgDir string
	for _, p := range pkgs {
		if p.ImportPath != pkg {
			continue
		}
		pkgDir = p.Dir
		break
	}
	if pkgDir == "" {
		return fmt.Errorf("package %s not found", pkg)
	}
	reverseFile := filepath.Join(pkgDir, file)

	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	overlay := map[string]interface{}{
		"Replace": map[string]string{
			reverseFile: absFile,
		},
	}
	overlayBytes, err := json.Marshal(overlay)
	if err != nil {
		panic(err)
	}
	return os.WriteFile(overlayFile, overlayBytes, 0755)
}
