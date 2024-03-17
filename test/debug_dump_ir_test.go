package test

import (
	"os"
	"testing"
)

// go test -run TestDumpIR -v ./test
func TestDumpIR(t *testing.T) {
	// must use "-a" or create a temp dir to ensure recompile
	// -a: cost 7~8s
	// give go.mod as a placeholder for go to build
	rootDir, tmpDir, err := tmpRuntimeModeAndTest("./testdata/dump_ir")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(rootDir)

	output, err := runXgo([]string{"--dump-ir", "main.Print", "--no-build-output", "--project-dir", tmpDir, "./"}, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// t.Logf("output:%s", output)
	seqs := []string{
		"DCLFUNC main.Print", // func decl
		"NAME-main.a",        // variable
		"CALLFUNC",           // call fmt.Printf
		"NAME-fmt.Printf",
		`LITERAL-"a:%s\n"`, // literal
		"CONVIFACE",        // convert a to interface{}
	}

	expectSequence(t, output, seqs)
}
