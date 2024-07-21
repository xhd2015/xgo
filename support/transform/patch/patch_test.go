package patch

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/assert"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goparse"
	"github.com/xhd2015/xgo/support/transform/astdiff"
)

func TestPatchFile(t *testing.T) {
	tests := []struct {
		dir string
	}{
		// {
		// 	dir: "./testdata/hello_world",
		// },
		// {
		// 	dir: "./testdata/replace",
		// },
		// {
		// 	dir: "./testdata/import",
		// },
		// {
		// 	dir: "./testdata/const_decl",
		// },
		{
			dir: "./testdata/prepend_func",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.dir, func(t *testing.T) {
			testPatchDir(t, tt.dir)
		})
	}
}

func testPatchDir(t *testing.T, dir string) {
	result, err := PatchFile(filepath.Join(dir, "original.go"), filepath.Join(dir, "original.go.patch"))
	if err != nil {
		t.Error(err)
		return
	}
	expectedBytes, err := fileutil.ReadFile(filepath.Join(dir, "expected.go"))
	if err != nil {
		t.Error(err)
		return
	}

	resultCode, _, err := goparse.ParseFileCode("result.go", []byte(result))
	if err != nil {
		t.Error(err)
		return
	}
	expectedCode, _, err := goparse.ParseFileCode("expected.go", expectedBytes)
	if err != nil {
		t.Error(err)
		return
	}

	// assert they are syntically the same
	if !astdiff.FileSame(resultCode, expectedCode) {
		result = strings.TrimSuffix(result, "\n")
		expected := strings.TrimSuffix(string(expectedBytes), "\n")
		if diff := assert.Diff(expected, result); diff != "" {
			t.Errorf("PatchFile(): %s", diff)
		}
	}
}
