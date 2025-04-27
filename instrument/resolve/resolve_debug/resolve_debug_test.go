package resolve_debug

import (
	"go/token"
	"testing"

	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/instrument_var"
	"github.com/xhd2015/xgo/instrument/load"
	"github.com/xhd2015/xgo/instrument/resolve"
)

func TestDebugResolve(t *testing.T) {
	t.Skip("local debug only")
	fset := token.NewFileSet()
	opts := load.LoadOptions{
		Dir:             "/path/to/secretflow/kuscia",
		FilterErrorFile: true,
		Fset:            fset,
	}
	packages := &edit.Packages{
		Fset:        fset,
		LoadOptions: opts,
	}
	err := packages.LoadPackages([]string{"./pkg/utils/paths"})
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
