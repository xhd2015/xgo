package instrument_compiler

import (
	"strings"
	"testing"
)

func TestNoderFiles_1_17_NoLeadingWhitespace(t *testing.T) {
	if len(NoderFiles_1_17) == 0 {
		t.Fatal("NoderFiles_1_17 is empty")
	}
	if strings.HasPrefix(NoderFiles_1_17, "\n") {
		t.Fatal("NoderFiles_1_17 starts with \\n")
	}
	if strings.HasPrefix(NoderFiles_1_17, " ") {
		t.Fatal("NoderFiles_1_17 starts with space")
	}
	if strings.HasPrefix(NoderFiles_1_17, "\t") {
		t.Fatal("NoderFiles_1_17 starts with tab")
	}
	if !strings.HasPrefix(NoderFiles_1_17, "// auto gen") {
		trimLen := 20
		if len(NoderFiles_1_17) < trimLen {
			trimLen = len(NoderFiles_1_17)
		}
		t.Fatalf("NoderFiles_1_17 should start with '// auto gen', got: %q", NoderFiles_1_17[:trimLen])
	}
}

func TestNoderFiles_1_20_NoLeadingWhitespace(t *testing.T) {
	if len(NoderFiles_1_20) == 0 {
		t.Fatal("NoderFiles_1_20 is empty")
	}
	if strings.HasPrefix(NoderFiles_1_20, "\n") {
		t.Fatal("NoderFiles_1_20 starts with \\n")
	}
	if strings.HasPrefix(NoderFiles_1_20, " ") {
		t.Fatal("NoderFiles_1_20 starts with space")
	}
	if strings.HasPrefix(NoderFiles_1_20, "\t") {
		t.Fatal("NoderFiles_1_20 starts with tab")
	}
	if !strings.HasPrefix(NoderFiles_1_20, "// auto gen") {
		trimLen := 20
		if len(NoderFiles_1_20) < trimLen {
			trimLen = len(NoderFiles_1_20)
		}
		t.Fatalf("NoderFiles_1_20 should start with '// auto gen', got: %q", NoderFiles_1_20[:trimLen])
	}
}

func TestNoderFiles_1_21_NoLeadingWhitespace(t *testing.T) {
	if len(NoderFiles_1_21) == 0 {
		t.Fatal("NoderFiles_1_21 is empty")
	}
	if strings.HasPrefix(NoderFiles_1_21, "\n") {
		t.Fatal("NoderFiles_1_21 starts with \\n")
	}
	if strings.HasPrefix(NoderFiles_1_21, " ") {
		t.Fatal("NoderFiles_1_21 starts with space")
	}
	if strings.HasPrefix(NoderFiles_1_21, "\t") {
		t.Fatal("NoderFiles_1_21 starts with tab")
	}
	if !strings.HasPrefix(NoderFiles_1_21, "// auto gen") {
		trimLen := 20
		if len(NoderFiles_1_21) < trimLen {
			trimLen = len(NoderFiles_1_21)
		}
		t.Fatalf("NoderFiles_1_21 should start with '// auto gen', got: %q", NoderFiles_1_21[:trimLen])
	}
}
