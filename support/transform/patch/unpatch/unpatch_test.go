// package unpatch removes all patch segments in a file
// restore a patched file to its original state,then
// apply patch again

package unpatch

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/assert"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goparse"
	"github.com/xhd2015/xgo/support/transform/astdiff"
)

func TestUnpatch(t *testing.T) {
	tests := []struct {
		dir     string
		wantErr error
	}{
		{
			dir: "simple",
		},
		{
			dir: "duplicate",
		},
		{
			dir:     "missing_close",
			wantErr: fmt.Errorf("missing close for /*<begin X*/"),
		},
		{
			dir:     "missing_id",
			wantErr: fmt.Errorf("missing id for /*<begin >*/"),
		},
		{
			dir:     "missing_end",
			wantErr: fmt.Errorf("missing /*<end X>*/"),
		},
		{
			dir: "recover_replaced",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.dir, func(t *testing.T) {
			text, err := fileutil.ReadFile(filepath.Join("./testdata", tt.dir, "original.go"))
			if err != nil {
				t.Error(err)
				return
			}
			got, err := Unpatch(string(text))
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("want err: %v, actual nil", tt.wantErr)
					return
				}
				wantErrMsg := tt.wantErr.Error()
				errMsg := err.Error()
				if !strings.Contains(errMsg, wantErrMsg) {
					t.Errorf("want err: %v, actual %s", wantErrMsg, errMsg)
					return
				}
				return
			}
			if err != nil {
				t.Error(err)
				return
			}
			want, err := fileutil.ReadFile(filepath.Join("./testdata", tt.dir, "expect.go"))
			if err != nil {
				t.Error(err)
				return
			}

			gotCode, _, err := goparse.ParseFileCode("result.go", []byte(got))
			if err != nil {
				t.Error(err)
				return
			}
			expectedCode, _, err := goparse.ParseFileCode("expected.go", want)
			if err != nil {
				t.Error(err)
				return
			}

			if !astdiff.FileSame(expectedCode, gotCode) {
				wantStr := string(want)
				t.Errorf("Unpatch() = %s", assert.Diff(wantStr, got))
			}
		})
	}
}
