package instrument_runtime

import (
	"strings"
	"testing"
)

// Representative newproc1 race-setup snippets (not full runtime files).
// Shape mirrors go1.18 / mid-1.19 / 1.20+ as used by buildRaceSetupAfterCreateCallback.
const (
	snippetGo118 = `
func newproc1(fn *funcval, callergp *g, callerpc uintptr) *g {
	// ... body ...
	if raceenabled {
		newg.racectx = racegostart(callerpc)
	}
	return newg
}

func saveAncestors(callergp *g) *[]ancestorInfo {
	// decoy: must not affect detection of raceignore / labels release
	newg.raceignore = 0
	racereleasemergeg(newg, unsafe.Pointer(&labelSync))
	return nil
}
`

	snippetGo119Early = `
func newproc1(fn *funcval, callergp *g, callerpc uintptr) *g {
	if raceenabled {
		newg.racectx = racegostart(callerpc)
		// early 1.19: labels release but no raceignore field write
		if newg.labels != nil {
			racereleasemergeg(newg, unsafe.Pointer(&labelSync))
		}
	}
	return newg
}

func other() {
	newg.raceignore = 0 // decoy after newproc1
}
`

	snippetGo120Plus = `
func newproc1(fn *funcval, callergp *g, callerpc uintptr) *g {
	// Set up race context.
	if raceenabled {
		newg.racectx = racegostart(callerpc)
		newg.raceignore = 0
		if newg.labels != nil {
			// See note in proflabel.go on labelSync's role
			racereleasemergeg(newg, unsafe.Pointer(&labelSync))
		}
	}
	return newg
}

func saveAncestors(callergp *g) *[]ancestorInfo {
	return nil
}
`
)

func TestBuildRaceSetupAfterCreateCallback(t *testing.T) {
	tests := []struct {
		name    string
		snippet string
		want    string
	}{
		{
			name:    "go1.18 racegostart only",
			snippet: snippetGo118,
			want:    "if raceenabled { newg.racectx = racegostart(pc); };",
		},
		{
			name:    "go1.19 early labels without raceignore",
			snippet: snippetGo119Early,
			want:    "if raceenabled { newg.racectx = racegostart(pc); if newg.labels != nil { racereleasemergeg(newg, unsafe.Pointer(&labelSync)) }; };",
		},
		{
			name:    "go1.20+ full block",
			snippet: snippetGo120Plus,
			want:    "if raceenabled { newg.racectx = racegostart(pc); newg.raceignore = 0; if newg.labels != nil { racereleasemergeg(newg, unsafe.Pointer(&labelSync)) }; };",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildRaceSetupAfterCreateCallback(tt.snippet)
			if got != tt.want {
				t.Fatalf("setup mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestContainsInFuncBoundsToFunctionBody(t *testing.T) {
	// Decoy after newproc1 must not count.
	if containsInFunc(snippetGo118, "func newproc1(", "newg.raceignore") {
		t.Fatal("expected raceignore decoy after newproc1 to be ignored")
	}
	if containsInFunc(snippetGo118, "func newproc1(", "racereleasemergeg(newg") {
		t.Fatal("expected racereleasemergeg decoy after newproc1 to be ignored")
	}
	if !containsInFunc(snippetGo120Plus, "func newproc1(", "newg.raceignore") {
		t.Fatal("expected raceignore inside newproc1 body")
	}
	if containsInFunc("func other() {}", "func newproc1(", "racegostart") {
		t.Fatal("missing start mark should be false")
	}
}

func TestContainsInFuncEOFWhenNoNextFunc(t *testing.T) {
	only := "func newproc1() {\n\tnewg.raceignore = 0\n}\n"
	// trailing content without \nfunc
	only = strings.TrimSuffix(only, "\n") + "\n// end of file\n"
	if !containsInFunc(only, "func newproc1(", "newg.raceignore") {
		t.Fatal("expected match when no next func and needle in body")
	}
}
