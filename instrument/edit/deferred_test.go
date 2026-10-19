package edit

import (
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/xgo/instrument/load"
	"github.com/xhd2015/xgo/support/goinfo"
)

func writeDeferredTestPkg(t *testing.T, dir string, name string) *load.Package {
	t.Helper()
	pkgDir := filepath.Join(dir, name)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, name+".go"), []byte("package "+name+"\n\nfunc F() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return &load.Package{
		GoPackage: &goinfo.Package{ImportPath: "example.com/" + name, Dir: pkgDir, GoFiles: []string{name + ".go"}},
		Deferred:  true,
	}
}

func TestDeferredPackages(t *testing.T) {
	dir := t.TempDir()
	fset := token.NewFileSet()
	// LoadOptions.Fset deliberately left nil: deferred parsing must use the
	// FileSet the packages are registered with
	c := &Packages{Fset: fset}
	a := writeDeferredTestPkg(t, dir, "a")
	b := writeDeferredTestPkg(t, dir, "b")
	c.Add(&load.Packages{Fset: fset, Packages: []*load.Package{a, b}})

	if len(c.Packages) != 0 || len(c.PackageByPath) != 0 {
		t.Fatalf("deferred packages must not be visible yet")
	}

	pkg, _, err := c.LoadPackage("example.com/a")
	if err != nil {
		t.Fatal(err)
	}
	if pkg == nil || len(pkg.Files) != 1 || pkg.Files[0].File.Syntax == nil {
		t.Fatalf("deferred package not parsed on demand: %+v", pkg)
	}
	if got := fset.Position(pkg.Files[0].File.Syntax.Package).Filename; got != filepath.Join(dir, "a", "a.go") {
		t.Errorf("position resolved against the wrong FileSet: %q", got)
	}

	// a merged package supersedes a deferred one
	override := &Packages{Fset: fset}
	merged := writeDeferredTestPkg(t, dir, "b")
	merged.Deferred = false
	load.ParseDeferred([]*load.Package{{GoPackage: merged.GoPackage, Deferred: true}}, load.LoadOptions{Fset: fset})
	override.Add(&load.Packages{Fset: fset, Packages: []*load.Package{{GoPackage: merged.GoPackage}}})
	c.Merge(override, true)
	want := c.PackageByPath["example.com/b"]
	c.ParseDeferred()
	if c.PackageByPath["example.com/b"] != want {
		t.Errorf("ParseDeferred replaced the merged package")
	}
	if len(c.Packages) != 2 {
		t.Errorf("packages: %d", len(c.Packages))
	}
}

func TestParseDeferredWithoutFset(t *testing.T) {
	pkg := writeDeferredTestPkg(t, t.TempDir(), "c")
	load.ParseDeferred([]*load.Package{pkg}, load.LoadOptions{})
	if pkg.Deferred || len(pkg.Files) != 1 || pkg.Files[0].Syntax == nil {
		t.Fatalf("not parsed: %+v", pkg)
	}
}
