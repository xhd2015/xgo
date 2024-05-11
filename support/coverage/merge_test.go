package coverage

import (
	"testing"
)

func mergeAndFormat(list ...string) string {
	var covList [][]*CovLine
	for _, content := range list {
		_, lines := Parse(content)
		covList = append(covList, lines)
	}
	lines := Merge(covList...)
	return Format("set", lines)
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name string
		covs []string
		want string
	}{
		{
			name: "empty",
			want: "mode: set",
		},
		{
			name: "single",
			covs: []string{`mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 0`},
			want: `mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 0`,
		},
		{
			name: "single_compact",
			covs: []string{`mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 1
github.com/xhd2015/xgo/runtime/core/func1.go:44.41,45.22 1 0
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 4`},
			want: `mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 5
github.com/xhd2015/xgo/runtime/core/func1.go:44.41,45.22 1 0`,
		},
		{
			name: "multiple_compact",
			covs: []string{`mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 1
github.com/xhd2015/xgo/runtime/core/func1.go:44.41,45.22 1 0
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 4`,
				`mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 0
github.com/xhd2015/xgo/runtime/core/func1.go:44.41,45.22 1 1
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 8`},
			want: `mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 13
github.com/xhd2015/xgo/runtime/core/func1.go:44.41,45.22 1 1`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeAndFormat(tt.covs...); got != tt.want {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}
