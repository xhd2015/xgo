package patch

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

func TestUpdateContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		beginMark   string
		endMark     string
		seq         []string
		i           int
		position    UpdatePosition
		addContent  string
		expected    string
		expectPanic bool
	}{
		{
			name: "add before sequence",
			content: `package main

func main() {
	fmt.Println("Hello")
}`,
			beginMark:  "/*<begin>*/",
			endMark:    "/*<end>*/",
			seq:        []string{"func main"},
			i:          0,
			position:   UpdatePosition_Before,
			addContent: `// This is added before main`,
			expected: `package main

/*<begin>*/// This is added before main/*<end>*/func main() {
	fmt.Println("Hello")
}`,
		},
		{
			name: "add after sequence",
			content: `package main

func main() {
	fmt.Println("Hello")
}`,
			beginMark:  "/*<begin>*/",
			endMark:    "/*<end>*/",
			seq:        []string{"func main", "(", ")", "{"},
			i:          3,
			position:   UpdatePosition_After,
			addContent: `	// This is added after main declaration`,
			expected: `package main

func main() {/*<begin>*/	// This is added after main declaration/*<end>*/
	fmt.Println("Hello")
}`,
		},
		{
			name: "replace existing content",
			content: `package main

func main() {/*<begin>*/	// Old comment/*<end>*/
	fmt.Println("Hello")
}`,
			beginMark:  "/*<begin>*/",
			endMark:    "/*<end>*/",
			seq:        []string{"func main", "(", ")", "{"},
			i:          3,
			position:   UpdatePosition_After,
			addContent: `	// New comment`,
			expected: `package main

func main() {/*<begin>*/	// New comment/*<end>*/
	fmt.Println("Hello")
}`,
		},
		{
			name: "missing sequence should panic",
			content: `package main

func main() {
	fmt.Println("Hello")
}`,
			beginMark:   "/*<begin>*/",
			endMark:     "/*<end>*/",
			seq:         []string{"func notExist"},
			i:           0,
			position:    UpdatePosition_Before,
			addContent:  `// This will panic`,
			expectPanic: true,
		},
		{
			name: "duplicate sequence should panic",
			content: `package main

func main() {
	fmt.Println("Hello")
}

func main() {
	fmt.Println("Duplicate")
}`,
			beginMark:   "/*<begin>*/",
			endMark:     "/*<end>*/",
			seq:         []string{"func main"},
			i:           0,
			position:    UpdatePosition_Before,
			addContent:  `// This will panic`,
			expectPanic: true,
		},
		{
			name: "empty add content",
			content: `package main

func main() {
	fmt.Println("Hello")
}`,
			beginMark:  "/*<begin>*/",
			endMark:    "/*<end>*/",
			seq:        []string{"func main"},
			i:          0,
			position:   UpdatePosition_Before,
			addContent: ``,
			expected: `package main

func main() {
	fmt.Println("Hello")
}`,
		},
		{
			name: "add content with special characters",
			content: `package main

func main() {
	fmt.Println("Hello")
}`,
			beginMark:  "/*<begin>*/",
			endMark:    "/*<end>*/",
			seq:        []string{"fmt.Println"},
			i:          0,
			position:   UpdatePosition_Before,
			addContent: `	log.Printf("Debug: %s", time.Now())`,
			expected: `package main

func main() {
	/*<begin>*/	log.Printf("Debug: %s", time.Now())/*<end>*/fmt.Println("Hello")
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but got none")
					}
				}()
			}

			result := UpdateContent(tt.content, tt.beginMark, tt.endMark, tt.seq, tt.i, tt.position, tt.addContent)

			if tt.expectPanic {
				t.Errorf("Expected panic but got none")
				return
			}

			// Normalize line endings for comparison
			result = strings.ReplaceAll(result, "\r\n", "\n")
			expected := strings.ReplaceAll(tt.expected, "\r\n", "\n")

			if result != expected {
				t.Errorf("UpdateContent() diff: %s", assert.Diff(expected, result))
			}
		})
	}
}

// Additional test for empty sequence
func TestUpdateContentNoSeparator_EmptySequence(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic with empty sequence but got none")
		}
	}()

	UpdateContent("content", "/*<begin>*/", "/*<end>*/", []string{}, 0, UpdatePosition_Before, "add")
}

// Test for out of bounds index
func TestUpdateContentNoSeparator_OutOfBoundsIndex(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic with out of bounds index but got none")
		}
	}()

	UpdateContent("content", "/*<begin>*/", "/*<end>*/", []string{"content"}, 1, UpdatePosition_Before, "add")
}

// Test for comparing UpdateContent vs UpdateContentLines difference
func TestUpdateContent_VsUpdateContentLines(t *testing.T) {
	content := "func main() {\n\tfmt.Println(\"test\")\n}"
	beginMark := "/*<begin>*/"
	endMark := "/*<end>*/"
	seq := []string{"func main"}
	addContent := "// Added content"

	// With UpdateContent (no separator)
	resultNoSeparator := UpdateContent(content, beginMark, endMark, seq, 0, UpdatePosition_Before, addContent)

	// With UpdateContentLines (newline separator)
	resultWithSeparator := UpdateContentLines(content, beginMark, endMark, seq, 0, UpdatePosition_Before, addContent)

	// Expected results
	expectedNoSeparator := "/*<begin>*/// Added content/*<end>*/func main() {\n\tfmt.Println(\"test\")\n}"
	expectedWithSeparator := "/*<begin>*/\n// Added content\n/*<end>*/\nfunc main() {\n\tfmt.Println(\"test\")\n}"

	// Normalize line endings for comparison
	resultNoSeparator = strings.ReplaceAll(resultNoSeparator, "\r\n", "\n")
	resultWithSeparator = strings.ReplaceAll(resultWithSeparator, "\r\n", "\n")

	if resultNoSeparator != expectedNoSeparator {
		t.Errorf("UpdateContent() diff: %s", assert.Diff(expectedNoSeparator, resultNoSeparator))
	}

	if resultWithSeparator != expectedWithSeparator {
		t.Errorf("UpdateContentLines() diff: %s", assert.Diff(expectedWithSeparator, resultWithSeparator))
	}
}
