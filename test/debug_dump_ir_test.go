package test

import (
	"os"
	"testing"
)

// go test -run TestDumpIR -v ./test
func TestDumpIR(t *testing.T) {
	t.Parallel()
	// must use "-a" or create a temp dir to ensure recompile
	// -a: cost 7~8s
	// give go.mod as a placeholder for go to build
	rootDir, tmpDir, err := tmpWithRuntimeGoModeAndTest("./testdata/dump_ir")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(rootDir)

	output, err := runXgo([]string{"--dump-ir", "main.Print", "--no-build-output", "--project-dir", tmpDir, "./"}, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// t.Logf("output:%s", output)

	goVersion, err := getGoVersion()
	if err != nil {
		t.Fatal(getErrMsg(err))
	}
	var seqs []string
	if goVersion.Major == 1 && goVersion.Minor <= 17 {
		seqs = append(seqs, "DCLFUNC", "FUNC-func(string)") // func decl
	} else {
		seqs = append(seqs, "DCLFUNC main.Print") // func decl
	}

	seqs = append(seqs,
		"NAME-main.a", // variable
		"CALLFUNC",    // call fmt.Printf
		"NAME-fmt.Printf",
		`LITERAL-"a:%s\n"`, // literal
		"CONVIFACE",        // convert a to interface{}
	)

	expectSequence(t, output, seqs)
}
