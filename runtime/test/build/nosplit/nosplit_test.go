package nosplit

import (
	"testing"
)

func TestShouldCompile(t *testing.T) {
	// should compile
}

func TestNoEscape(t *testing.T) {
	noescape(nil)
	// should build pass
}

func TestNoEscapeSmall(t *testing.T) {
	nosplitSmall(nil)
	// should build pass
}
