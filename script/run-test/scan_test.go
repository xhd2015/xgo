package main

import (
	"reflect"
	"testing"
)

func TestSplitList(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty string",
			content:  "",
			expected: nil,
		},
		{
			name:     "single word",
			content:  "word",
			expected: []string{"word"},
		},
		{
			name:     "multiple words",
			content:  "one two three",
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "leading spaces",
			content:  "  leading spaces",
			expected: []string{"leading", "spaces"},
		},
		{
			name:     "trailing spaces",
			content:  "trailing spaces  ",
			expected: []string{"trailing", "spaces"},
		},
		{
			name:     "multiple spaces between words",
			content:  "multiple  spaces   between    words",
			expected: []string{"multiple", "spaces", "between", "words"},
		},
		{
			name:     "escaped spaces",
			content:  "escaped\\ space",
			expected: []string{"escaped space"},
		},
		{
			name:     "multiple escaped spaces",
			content:  "multiple\\ escaped\\ spaces",
			expected: []string{"multiple escaped spaces"},
		},
		{
			name:     "mixed regular and escaped spaces",
			content:  "regular and\\ escaped spaces",
			expected: []string{"regular", "and escaped", "spaces"},
		},
		{
			name:     "escaped space at end",
			content:  "escaped at end\\ ",
			expected: []string{"escaped", "at", "end "},
		},
		{
			name:     "backslash not followed by space",
			content:  "backslash\\not\\followed\\by\\space",
			expected: []string{"backslash\\not\\followed\\by\\space"},
		},
		{
			name:     "gcflags parsed",
			content:  `-gcflags=all=-N\ -l -v`,
			expected: []string{"-gcflags=all=-N -l", "-v"},
		},
		{
			name:     "multiple spaces trimmed parsed",
			content:  `   -a    -b  -c `,
			expected: []string{"-a", "-b", "-c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitList(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("splitList(%q) = %v, expected %v", tt.content, result, tt.expected)
			}
		})
	}
}
