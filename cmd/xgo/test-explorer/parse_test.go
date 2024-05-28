package test_explorer

import (
	"strings"
	"testing"
)

func TestParseTestFuncsCode(t *testing.T) {
	tests := []struct {
		name  string
		test  string
		code  string
		flags string
		args  string
		err   string
	}{
		{
			name: "arg with spaces",
			test: "TestFunc",
			code: `package test
			import "testing"

			// args:   -arg    ${PROJECT_DIR}/x
			// flags: -p   1
			func TestFunc(t *testing.T){}
			`,
			args:  "-arg CWD/x",
			flags: "-p 1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, funcDecls, err := parseTestFuncsCode("test.go", strings.NewReader(tt.code))
			var actualErr string
			if err != nil {
				actualErr = err.Error()
			}
			if actualErr != tt.err {
				t.Errorf("parseTestFuncsCode() error = %v, wantErr %v", err, tt.err)
				return
			}
			fn, err := getFuncDecl(funcDecls, tt.test)
			if err != nil {
				t.Fatal(err)
			}
			flags, args, err := parseFuncArgs(fn)
			if err != nil {
				t.Fatal(err)
			}
			args = applyVars("CWD", args)
			flags = applyVars("CWD", flags)

			flagStr := strings.Join(flags, " ")
			if tt.flags != flagStr {
				t.Errorf("parseTestFuncsCode() flags = %v, want: %v", flagStr, tt.flags)
				return
			}

			argStr := strings.Join(args, " ")
			if tt.args != argStr {
				t.Errorf("parseTestFuncsCode() args = %v, want: %v", argStr, tt.args)
				return
			}
		})
	}
}
