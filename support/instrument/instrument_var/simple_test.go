package instrument_var_test

import (
	"os"
	"testing"

	"github.com/xhd2015/xgo/support/instrument/edit"
	"github.com/xhd2015/xgo/support/instrument/instrument_var"
	"github.com/xhd2015/xgo/support/instrument/load"
)

func TestSimple(t *testing.T) {
	loadPackages, err := load.LoadPackages([]string{"./testdata/simple/..."}, load.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	packages := edit.Edit(loadPackages)
	instrument_var.Instrument(packages)

	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			if !file.HasEdit() {
				t.Logf("file unchanged: %s", file.File.Name)
				continue
			}
			edit := file.Edit
			t.Logf("file changed: %s", file.File.Name)
			t.Logf("content: %s", edit.String())
		}
	}
}

func TestVarGroup(t *testing.T) {
	loadPackages, err := load.LoadPackages([]string{"./testdata/var_group/..."}, load.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	packages := edit.Edit(loadPackages)
	instrument_var.Instrument(packages)
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			if !file.HasEdit() {
				t.Logf("file unchanged: %s", file.File.Name)
				continue
			}
			edit := file.Edit
			t.Logf("file changed: %s", file.File.Name)
			t.Logf("content: %s", edit.String())
		}
	}
}

func TestCustom(t *testing.T) {
	dir := os.Getenv("CUSTOM_TEST_DIR")
	if dir == "" {
		t.Skip("CUSTOM_TEST_DIR is not set")
	}
	args := []string{"./..."}
	mod := "vendor"
	loadPackages, err := load.LoadPackages(args, load.LoadOptions{
		Dir: dir,
		Mod: mod,
	})
	if err != nil {
		t.Fatal(err)
	}
	packages := edit.Edit(loadPackages)
	instrument_var.Instrument(packages)
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			if !file.HasEdit() {
				t.Logf("file unchanged: %s", file.File.Name)
				continue
			}
			edit := file.Edit
			t.Logf("file changed: %s", file.File.Name)
			t.Logf("content: %s", edit.String())
		}
	}
}
