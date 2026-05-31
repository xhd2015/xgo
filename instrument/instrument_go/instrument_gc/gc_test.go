package instrument_gc

import (
	"strings"
	"testing"
)

func TestInterceptCompile_NoLeadingWhitespace(t *testing.T) {
	if len(interceptCompile) == 0 {
		t.Fatal("interceptCompile is empty")
	}
	if strings.HasPrefix(interceptCompile, "\n") {
		t.Fatal("interceptCompile starts with \\n")
	}
	if strings.HasPrefix(interceptCompile, " ") {
		t.Fatal("interceptCompile starts with space")
	}
	if strings.HasPrefix(interceptCompile, "\t") {
		t.Fatal("interceptCompile starts with tab")
	}
	if !strings.HasPrefix(interceptCompile, "var dlvDebug bool") {
		trimLen := 50
		if len(interceptCompile) < trimLen {
			trimLen = len(interceptCompile)
		}
		t.Fatalf("interceptCompile should start with 'var dlvDebug bool', got: %q", interceptCompile[:trimLen])
	}
}

func TestPipeoutput_NoLeadingWhitespace(t *testing.T) {
	if len(pipeoutput) == 0 {
		t.Fatal("pipeoutput is empty")
	}
	if strings.HasPrefix(pipeoutput, "\n") {
		t.Fatal("pipeoutput starts with \\n")
	}
	if strings.HasPrefix(pipeoutput, " ") {
		t.Fatal("pipeoutput starts with space")
	}
	if strings.HasPrefix(pipeoutput, "\t") {
		t.Fatal("pipeoutput starts with tab")
	}
	if !strings.HasPrefix(pipeoutput, "if dlvDebug") {
		trimLen := 30
		if len(pipeoutput) < trimLen {
			trimLen = len(pipeoutput)
		}
		t.Fatalf("pipeoutput should start with 'if dlvDebug', got: %q", pipeoutput[:trimLen])
	}
}
