package test

import (
	"os"
	"testing"
)

// go test -run TestDumpVscode -v ./test
func TestDumpVscode(t *testing.T) {
	rootDir, tmpDir, err := tmpRuntimeModeAndTest("./testdata/dump_ir")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(rootDir)

	output, err := xgoBuild([]string{"--debug", "main", "--vscode", "stdout?nowait", "--no-build-output",
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
		`"-nolocalimports",`,
		`"-importcfg",`,
		`"-pack",`,

		// env
		`"env": {`,
		`"XGO_COMPILER_ENABLE": "true",`,

		`"mode": "exec",`,
		`"request": "launch",`,
		`"type": "go"`,
	}
	expectSequence(t, output, seqs)
}
