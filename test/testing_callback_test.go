package test

import (
	"testing"
)

// go test -run TestTestingCallbackShouldBeCalled -v ./test
func TestTestingCallbackShouldBeCalled(t *testing.T) {
	t.Parallel()
	testTrapWithTest(t, "./testdata/testing_callback", func(output string) error {
		// t.Logf("orig output: %s", output)
		expectSequence(t, output, []string{
			"RUN", "TestExample\n",
			"main_test.go", "TEST EXAMPLE\n",
			"PASS",
		})
		return nil
	}, func(output string) error {
		// t.Logf("instrument output: %s", output)
		// TEST STARTED appears twice, first without name, then with name
		expectSequence(t, output, []string{
			"TEST STARTED: \n",
			"RUN", "TestExample\n",
			"TEST STARTED: TestExample\n",
			"main_test.go", "TEST EXAMPLE\n",
			"PASS",
		})
		return nil
	})
}
