package patch

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

func TestUpdateContentLines(t *testing.T) {
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

/*<begin>*/
// This is added before main
/*<end>*/
func main() {
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
			seq:        []string{"func main", "(", ")", "{\n"},
			i:          3,
			position:   UpdatePosition_After,
			addContent: `	// This is added after main declaration`,
			expected: `package main

func main() {
/*<begin>*/
	// This is added after main declaration
/*<end>*/
	fmt.Println("Hello")
}`,
		},
		{
			name: "replace existing content",
			content: `package main

func main() {
/*<begin>*/
	// Old comment
/*<end>*/
	fmt.Println("Hello")
}`,
			beginMark:  "/*<begin>*/",
			endMark:    "/*<end>*/",
			seq:        []string{"func main", "(", ")", "{\n"},
			i:          3,
			position:   UpdatePosition_After,
			addContent: `	// New comment`,
			expected: `package main

func main() {
/*<begin>*/
	// New comment
/*<end>*/

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

			result := UpdateContentLines(tt.content, tt.beginMark, tt.endMark, tt.seq, tt.i, tt.position, tt.addContent)

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
func TestUpdateContent_EmptySequence(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic with empty sequence but got none")
		}
	}()

	UpdateContentLines("content", "/*<begin>*/", "/*<end>*/", []string{}, 0, UpdatePosition_Before, "add")
}

// Test for out of bounds index
func TestUpdateContent_OutOfBoundsIndex(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic with out of bounds index but got none")
		}
	}()

	UpdateContentLines("content", "/*<begin>*/", "/*<end>*/", []string{"content"}, 1, UpdatePosition_Before, "add")
}
