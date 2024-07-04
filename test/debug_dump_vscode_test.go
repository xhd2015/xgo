package test

import (
	"os"
	"testing"
)

// go test -run TestDumpVscode -v ./test
func TestDumpVscode(t *testing.T) {
	t.Parallel()
	goVersion, err := getGoVersion()
	if err != nil {
		t.Fatal(getErrMsg(err))
	}
	rootDir, tmpDir, err := tmpWithRuntimeGoModeAndTest("./testdata/dump_ir")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(rootDir)

	output, err := runXgo([]string{"--debug-target", "main", "--vscode", "stdout?nowait", "--no-build-output",
		// "-a",// debug
		"--project-dir", tmpDir, "./"}, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// t.Logf("output:%s", output)
	seqs := []string{
		`"configurations": [`,

		// args
		`"args": [`,
		`"-trimpath",`,
		` "-buildid",`,
	}
	if goVersion.Major == 1 && goVersion.Minor >= 18 {
		seqs = append(seqs, `"-nolocalimports",`) // not appear at go1.17
	}
	seqs = append(seqs,
		`"-importcfg",`,
		`"-pack",`,

		// env
		`"env": {`,
		`"XGO_COMPILER_ENABLE": "true",`,

		`"mode": "exec",`,
		`"request": "launch",`,
		`"type": "go"`,
	)
	expectSequence(t, output, seqs)
}
