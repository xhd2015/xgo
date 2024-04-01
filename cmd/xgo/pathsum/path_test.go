package pathsum

import "testing"

// go test -run TestProcessSpecial -v ./cmd/xgo/pathsum
func TestProcessSpecial(t *testing.T) {
	testCases := []*struct {
		Arg string
		Res string
	}{
		{"", ""},
		{"/", ""},
		{"/a", "a"},
		{"C:/a", "Ca"}, // Windows
		{"C:\\a", "Ca"},
		{"/ab/c", "abc"},
		{"/ab/c", "abc"},
		{"a的b", "a的b"}, // CN
	}

	for _, testCase := range testCases {
		res := processSpecial(testCase.Arg)
		if res != testCase.Res {
			t.Fatalf("expect processSpecial(%q) = %q, actual: %q", testCase.Arg, testCase.Res, res)
		}
	}
}
