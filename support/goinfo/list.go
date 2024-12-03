package goinfo

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

// go list ./pkg
func ListPackages(dir string, mod string, args []string) ([]string, error) {
	flags := []string{"list"}
	if mod != "" {
		flags = append(flags, "-mod="+mod)
	}
	flags = append(flags, args...)
	output, err := cmd.Dir(dir).Output("go", flags...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(output, "\n")
	return lines, nil
}

// ListRelativeFiles list all go files, return a list of relative paths
func ListRelativeFiles(dir string, args []string) ([]string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	flags := []string{"list", "-e", "-json"}
	flags = append(flags, args...)
	res, err := cmd.Dir(dir).Output("go", flags...)
	if err != nil {
		return nil, err
	}
	var relFiles []string
	dec := json.NewDecoder(strings.NewReader(res))
	for dec.More() {
		// check 'go help list'
		var pkg struct {
			// the abs dir
			Dir string
			// the file names
			GoFiles     []string
			TestGoFiles []string
		}
		err := dec.Decode(&pkg)
		if err != nil {
			return nil, err
		}
		goFiles := make([]string, 0, len(pkg.GoFiles)+len(pkg.TestGoFiles))
		goFiles = append(goFiles, pkg.GoFiles...)
		goFiles = append(goFiles, pkg.TestGoFiles...)
		if len(goFiles) > 0 {
			absPkgDir, err := filepath.Abs(pkg.Dir)
			if err != nil {
				return nil, err
			}
			for _, goFile := range goFiles {
				if !strings.HasSuffix(goFile, ".go") {
					// some cache files
					continue
				}
				absGoFile := goFile
				if !filepath.IsAbs(goFile) {
					// go list outputs file names
					// in most cases this goFile is not absolute
					absGoFile = filepath.Join(absPkgDir, goFile)
				}

				// this is just for compatibility
				if !strings.HasPrefix(absGoFile, absDir) {
					continue
				}
				relFile := strings.TrimPrefix(absGoFile[len(absDir):], string(filepath.Separator))
				if relFile != "" {
					relFiles = append(relFiles, relFile)
				}
			}
		}
	}
	return relFiles, nil
}
