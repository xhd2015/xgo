package main

import (
	"testing"
)

func TestAddBlankImport(t *testing.T) {
	importStr := `import _ "github.com/xhd2015/xgo/runtime/trace"`

	tests := []struct {
		name    string
		content string
		want    string
		wantOK  bool
	}{
		{
			name:    "simple package line",
			content: "package pkg\n",
			want:    "package pkg;" + importStr + "\n",
			wantOK:  true,
		},
		{
			name:    "package line without newline",
			content: "package pkg",
			want:    "package pkg;" + importStr,
			wantOK:  true,
		},
		{
			name:    "package line already ends with semicolon",
			content: "package pkg;import \"runtime\";\n",
			want:    "package pkg;import \"runtime\";" + importStr + "\n",
			wantOK:  true,
		},
		{
			name:    "package line with no trailing semicolon",
			content: "package pkg\n",
			want:    "package pkg;" + importStr + "\n",
			wantOK:  true,
		},
		{
			name:    "no package keyword",
			content: "not a go file",
			want:    "",
			wantOK:  false,
		},
		{
			name:    "package line ending with semicolon and \\r\\n",
			content: "package pkg;\r\n",
			want:    "package pkg;" + importStr + "\r\n",
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := addBlankImport(tt.content)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}
