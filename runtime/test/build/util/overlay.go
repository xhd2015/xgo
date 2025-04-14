package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/goinfo"
)

func GenerateOverlay(dir string, overlayFile string, pkg string, file string, replacedFile string) error {
	pkgs, err := goinfo.ListPackages([]string{pkg}, goinfo.LoadPackageOptions{
		Dir: dir,
	})
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("package %s not found", pkg)
	}
	var foundPkg *goinfo.Package
	var pkgDir string
	if len(pkgs) > 1 {
		for _, p := range pkgs {
			if p.ImportPath != pkg {
				continue
			}
			foundPkg = p
			pkgDir = p.Dir
			break
		}
	} else {
		foundPkg = pkgs[0]
		pkgDir = foundPkg.Dir
	}
	if pkgDir == "" {
		return fmt.Errorf("package %s not found", pkg)
	}
	var foundFile bool
	for _, f := range foundPkg.GoFiles {
		if f == file {
			foundFile = true
			break
		}
	}
	if !foundFile {
		return fmt.Errorf("file %s not found", file)
	}
	origFile := filepath.Join(pkgDir, file)

	absFile, err := filepath.Abs(replacedFile)
	if err != nil {
		return err
	}
	overlay := map[string]interface{}{
		"Replace": map[string]string{
			origFile: absFile,
		},
	}
	overlayBytes, err := json.Marshal(overlay)
	if err != nil {
		panic(err)
	}
	return os.WriteFile(overlayFile, overlayBytes, 0755)
}
