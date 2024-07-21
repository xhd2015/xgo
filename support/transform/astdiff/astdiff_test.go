package astdiff

import (
	"testing"

	"github.com/xhd2015/xgo/support/goparse"
)

func TestNodeSame(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "empty",
			want: true,
		},
		{
			name: "func with comment",
			a:    `func main(){}`,
			b:    "func     main(){ /**/}",
			want: true,
		},
		{
			name: "func with space",
			a:    `func main() string {}`,
			b: `func     main()    string { 
			}`,
			want: true,
		},
		{
			name: "func signature changed",
			a:    `func greet(s string){}`,
			b:    "func     greet(s string,version int)  { return \"\" }",
			want: false,
		},
		{
			name: "return statement",
			a:    `func greet(s string) string{ return "hello " + s }`,
			b: `func     greet(s string,)  string{ 
			return  "hello " + s
			}`,
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testNodeSame(t, tt.a, tt.b, tt.want)
		})
	}
}

func testNodeSame(t *testing.T, a string, b string, want bool) {
	fileA, _, err := goparse.ParseFileCode("a.go", []byte(goparse.AddMissingPackage(a, "main")))
	if err != nil {
		t.Error(err)
		return
	}
	fileB, _, err := goparse.ParseFileCode("b.go", []byte(goparse.AddMissingPackage(b, "main")))
	if err != nil {
		t.Error(err)
		return
	}
	got := NodeSame(fileA, fileB)
	if got != want {
		t.Errorf("expect node same: %v, actual: %v", want, got)
	}
}
