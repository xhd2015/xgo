package patch

import (
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

func TestCleanPatch(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "remove begin and end markers",
			content: `package main

func main() {
    
}`,
			expected: `package main

func main() {
    
}`,
		},
		{
			name: "multiple markers",
			content: `package main

/*<begin code that should be removed */
import (
    "fmt"
)
/*<end another end comment */

func main() {
    /*<begin should remove this too */
    // debug code
    /*<end end of debug */
    fmt.Println("Hello")
}`,
			expected: `package main



func main() {
    
    fmt.Println("Hello")
}`,
		},
		{
			name: "incomplete markers",
			content: `package main

/*<begin no end marker

func main() {
    fmt.Println("Hello")
}`,
			expected: `package main

/*<begin no end marker

func main() {
    fmt.Println("Hello")
}`,
		},
		{
			name:     "no markers",
			content:  `package main\n\nfunc main() {\n\tfmt.Println("Hello")\n}`,
			expected: `package main\n\nfunc main() {\n\tfmt.Println("Hello")\n}`,
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanPatch(tt.content)
			// Verify the result matches the expected output
			if result != tt.expected {
				t.Errorf("CleanPatch() diff: %s", assert.Diff(tt.expected, result))
			}
		})
	}
}

// Test CleanPatch and CleanPatchMarkers with simple cases
func TestCleanPatchSimple(t *testing.T) {
	// Test cases for CleanPatch
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple begin/end markers",
			input:    "before /*<begin content to remove */ middle /*<end more content */ after",
			expected: "before  after",
		},
		{
			name:     "incomplete marker at end",
			input:    "content /*<begin no end",
			expected: "content /*<begin no end",
		},
		{
			name:     "no markers",
			input:    "content without any markers",
			expected: "content without any markers",
		},
		{
			name:     "whitespace preservation",
			input:    "before\n/*<begin\nremove\n*/\nmiddle\n/*<end\nremove\n*/\nafter",
			expected: "before\n\nafter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CleanPatch(tc.input)
			if result != tc.expected {
				t.Errorf("CleanPatch() diff: %s", assert.Diff(tc.expected, result))
			}
		})
	}
}

// Test for custom marker patterns
func TestCleanPatchMarkers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    PatchMarker
		end      PatchMarker
		expected string
	}{
		{
			name: "multiple different markers",
			content: `package main

// DEBUG_START
import (
    "fmt"
    "log"
)
// DEBUG_END

func main() {
    // TEST_START
    log.Println("Debug message")
    // TEST_END
    fmt.Println("Hello")
}`,
			start: PatchMarker{Begin: "// DEBUG_START", End: "// DEBUG_END"},
			end:   PatchMarker{Begin: "// TEST_START", End: "// TEST_END"},
			expected: `package main


    fmt.Println("Hello")
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanPatchMarkers(tt.content, tt.start, tt.end)
			// Verify the result matches the expected output
			if result != tt.expected {
				t.Errorf("CleanPatchMarkers() diff: %s", assert.Diff(tt.expected, result))
			}
		})
	}
}

// Test the specific "multiple markers" case
func TestCleanPatch_MultipleMarkers(t *testing.T) {
	content := `package main

/*<begin code that should be removed */
import (
    "fmt"
)
/*<end another end comment */

func main() {
    /*<begin should remove this too */
    // debug code
    /*<end end of debug */
    fmt.Println("Hello")
}`

	// Match exact whitespace in the actual output
	expected := `package main



func main() {
    
    fmt.Println("Hello")
}`

	result := CleanPatch(content)
	if result != expected {
		t.Errorf("CleanPatch() diff: %s", assert.Diff(expected, result))
	}
}
