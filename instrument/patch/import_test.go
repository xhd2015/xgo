package patch

import (
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

func TestAddImportAfterName(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		beginMark string
		endMark   string
		impName   string
		pkgPath   string
		want      string
	}{
		{
			name:      "basic import with name",
			code:      "package main;/*<begin>*/;/*<end>*/",
			beginMark: "/*<begin>*/",
			endMark:   "/*<end>*/",
			impName:   "fmt",
			pkgPath:   "fmt",
			want:      "package main;/*<begin>*/;import fmt \"fmt\"/*<end>*/",
		},
		{
			name:      "import without name",
			code:      "package main;/*<begin>*/;/*<end>*/",
			beginMark: "/*<begin>*/",
			endMark:   "/*<end>*/",
			impName:   "",
			pkgPath:   "fmt",
			want:      "package main;/*<begin>*/;import \"fmt\"/*<end>*/",
		},
		{
			name:      "import with existing package",
			code:      "package main;import \"os\";/*<begin>*/;/*<end>*/",
			beginMark: "/*<begin>*/",
			endMark:   "/*<end>*/",
			impName:   "fmt",
			pkgPath:   "fmt",
			want:      "package main;import \"os\";/*<begin>*/;import fmt \"fmt\"/*<end>*/",
		},
		{
			name:      "import with markers",
			code:      "package main;/*<begin>*/;/*<end>*/",
			beginMark: "/*<begin>*/",
			endMark:   "/*<end>*/",
			impName:   "fmt",
			pkgPath:   "fmt",
			want:      "package main;/*<begin>*/;import fmt \"fmt\"/*<end>*/",
		},
		{
			name:      "empty package path",
			code:      "package main;/*<begin>*/;/*<end>*/",
			beginMark: "/*<begin>*/",
			endMark:   "/*<end>*/",
			impName:   "fmt",
			pkgPath:   "",
			want:      "package main;/*<begin>*/;import fmt \"\"/*<end>*/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddImportAfterName(tt.code, tt.beginMark, tt.endMark, tt.impName, tt.pkgPath)
			if diff := assert.Diff(tt.want, got); diff != "" {
				t.Errorf("AddImportAfterName() diff: %s", diff)
			}
		})
	}
}
