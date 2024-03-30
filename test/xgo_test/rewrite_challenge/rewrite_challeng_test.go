package rewrite_challenge

import "testing"

func __debug_ir_rewrite(s string) string {
	panic("should be replaced by compiler")
}

func TestRewriteChallenge(t *testing.T) {
	text := "Hello IR"
	result := __debug_ir_rewrite(text)
	if text != result {
		t.Fatalf("expect result: %s, actual: %s", text, result)
	}
}
