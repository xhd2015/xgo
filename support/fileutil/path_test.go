package fileutil

import "testing"

// go test -run TestCleanSpecial -v ./cmd/xgo/pathsum
func TestCleanSpecial(t *testing.T) {
	testCases := []*struct {
		Arg string
		Res string
	}{
		{"", ""},
		{" ", ""}, // space to _
		{"/", ""},
		{"/a", "a"},
		{"C:/a", "Ca"}, // Windows
		{"C:\\a", "Ca"},
		{"/ab/c", "abc"},
		{"/ab/c", "abc"},
		{"a的b", "a的b"}, // CN
	}

	for _, testCase := range testCases {
		res := CleanSpecial(testCase.Arg)
		if res != testCase.Res {
			t.Fatalf("expect CleanSpecial(%q) = %q, actual: %q", testCase.Arg, testCase.Res, res)
		}
	}
}
