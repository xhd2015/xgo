package instrument_var_test

import (
	"go/token"
	"os"
	"testing"

	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/instrument_var"
	"github.com/xhd2015/xgo/instrument/load"
	"github.com/xhd2015/xgo/instrument/resolve"
)

func TestSimple(t *testing.T) {
	fset := token.NewFileSet()
	opts := load.LoadOptions{
		FilterErrorFile: true,
		Fset:            fset,
	}
	packages := &edit.Packages{
		Fset:        fset,
		LoadOptions: opts,
	}
	err := packages.LoadPackages([]string{"./testdata/simple/..."})
	if err != nil {
		t.Fatal(err)
	}
	registry := resolve.NewPackagesRegistry(packages)
	err = resolve.Traverse(registry, packages.Packages, &resolve.Recorder{})
	if err != nil {
		t.Fatal(err)
	}
	recorder := &resolve.Recorder{}
	err = instrument_var.TrapVariables(packages, recorder)
	if err != nil {
		t.Fatal(err)
	}

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
	fset := token.NewFileSet()
	opts := load.LoadOptions{
		FilterErrorFile: true,
		Fset:            fset,
	}
	packages := &edit.Packages{
		Fset:        fset,
		LoadOptions: opts,
	}
	err := packages.LoadPackages([]string{"./testdata/var_group/..."})
	if err != nil {
		t.Fatal(err)
	}
	registry := resolve.NewPackagesRegistry(packages)
	err = resolve.Traverse(registry, packages.Packages, &resolve.Recorder{})
	if err != nil {
		t.Fatal(err)
	}
	recorder := &resolve.Recorder{}
	err = instrument_var.TrapVariables(packages, recorder)
	if err != nil {
		t.Fatal(err)
	}
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
	fset := token.NewFileSet()
	opts := load.LoadOptions{
		FilterErrorFile: true,
		Fset:            fset,
		Mod:             "vendor",
	}
	packages := &edit.Packages{
		Fset:        fset,
		LoadOptions: opts,
	}
	dir := os.Getenv("CUSTOM_TEST_DIR")
	if dir == "" {
		t.Skip("CUSTOM_TEST_DIR is not set")
	}
	args := []string{"./..."}
	err := packages.LoadPackages(args)
	if err != nil {
		t.Fatal(err)
	}
	registry := resolve.NewPackagesRegistry(packages)
	err = resolve.Traverse(registry, packages.Packages, &resolve.Recorder{})
	if err != nil {
		t.Fatal(err)
	}
	recorder := &resolve.Recorder{}
	err = instrument_var.TrapVariables(packages, recorder)
	if err != nil {
		t.Fatal(err)
	}
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
