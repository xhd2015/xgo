//go:build go1.25

package as_unit_test_run

import "testing"

func TestInstrumentedGoRebuildCompileShouldPassWithBypassEnv(t *testing.T) {
	err := RunCommandInResolvedRootDir("go", "run", "./test/integrations/test_instrumented_go_rebuild_compile_shoud_pass_with_bypass_env")
	if err != nil {
		t.Fatal(err)
	}
}
