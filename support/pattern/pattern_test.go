package pattern

import "testing"

func TestMatch(t *testing.T) {
	patterns := CompilePatterns([]string{"a/b/c/**"})

	m := patterns.MatchAny("a/b/c/d/e")
	if !m {
		t.Fatalf("expect match, actual: %v", m)
	}
}
