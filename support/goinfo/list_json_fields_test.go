package goinfo

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Listing only the fields of Package must not change what is decoded.
func TestListPackagesJSONFields(t *testing.T) {
	out, err := GetGoVersionOutput("go")
	if err != nil {
		t.Skipf("go: %v", err)
	}
	goVersion, err := ParseGoVersion(out)
	if err != nil {
		t.Fatal(err)
	}
	if goVersion.Major == 1 && goVersion.Minor < 19 {
		t.Skipf("go list -json=fields needs go1.19: %s", out)
	}
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/m\n\ngo 1.18\n")
	write("a/a.go", "package a\n\nimport \"strings\"\n\nfunc A() string { return strings.ToUpper(\"a\") }\n")
	write("a/a_test.go", "package a\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n")
	write("a/a_ext_test.go", "package a_test\n\nimport (\n\t\"testing\"\n\n\t\"example.com/m/b\"\n)\n\nfunc TestB(t *testing.T) { b.B() }\n")
	write("b/b.go", "package b\n\nimport \"example.com/m/a\"\n\nfunc B() string { return a.A() }\n")
	write("onlytest/x_test.go", "package onlytest\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) {}\n")
	write("broken/broken.go", "package broken\n\nimport \"example.com/m/missing\"\n\nvar _ = missing.X\n")

	list := func(v *GoVersion) []*Package {
		t.Helper()
		pkgs, err := ListPackages([]string{"./..."}, LoadPackageOptions{Dir: dir, Deps: true, Test: true, GoVersion: v})
		if err != nil {
			t.Fatal(err)
		}
		return pkgs
	}
	all := list(nil)
	fields := list(goVersion)
	if len(all) < 10 {
		t.Fatalf("expected the module and its std deps, got %d packages", len(all))
	}
	if !reflect.DeepEqual(all, fields) {
		for i := range all {
			if i < len(fields) && !reflect.DeepEqual(all[i], fields[i]) {
				t.Fatalf("package %d differs:\n-json:        %+v\n-json=fields: %+v", i, *all[i], *fields[i])
			}
		}
		t.Fatalf("got %d packages with -json, %d with -json=fields", len(all), len(fields))
	}
}
