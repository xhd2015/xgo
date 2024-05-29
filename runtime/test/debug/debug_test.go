// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import "testing"

func TestHello(t *testing.T) {
	t.Logf("hello world!")
}
