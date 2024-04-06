package goinfo

import "testing"

func TestParseMode(t *testing.T) {
	testCases := []struct {
		Content string
		Module  string
	}{
		{
			`  module a`,
			`a`,
		},
		{
			`module    a/bc//yes it me`,
			`a/bc`,
		},
		{
			"go 1.18\r\nmodule    a/bc//windows it me\r\n",
			`a/bc`,
		},
	}
	for _, tc := range testCases {
		m := parseModPath(tc.Content)
		if m != tc.Module {
			t.Fatalf("expect parseModPath(%q) to be %q, actual: %q", tc.Content, tc.Module, m)
		}
	}
}
