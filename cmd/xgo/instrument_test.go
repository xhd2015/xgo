package main

import (
	"bytes"
	"testing"
)

func TestGitignoreAdd(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		want     []byte
		wantBool bool
	}{
		{
			name:     "empty content",
			content:  []byte(""),
			want:     []byte(XGO_GEN_IGNORE_PATTERN + "\n"),
			wantBool: true,
		},
		{
			name:     "existing content",
			content:  []byte("*.log\n"),
			want:     []byte("*.log\n" + XGO_GEN_IGNORE_PATTERN + "\n"),
			wantBool: true,
		},
		{
			name:     "pattern exists",
			content:  []byte(XGO_GEN_IGNORE_PATTERN + "\n"),
			want:     []byte(XGO_GEN_IGNORE_PATTERN + "\n"),
			wantBool: false,
		},
		{
			name:     "pattern exists at end",
			content:  []byte("*.log\n" + XGO_GEN_IGNORE_PATTERN),
			want:     []byte("*.log\n" + XGO_GEN_IGNORE_PATTERN),
			wantBool: false,
		},
		{
			name:     "pattern exists as whole",
			content:  []byte(XGO_GEN_IGNORE_PATTERN),
			want:     []byte(XGO_GEN_IGNORE_PATTERN),
			wantBool: false,
		},
		{
			name:     "pattern exists with trailing space",
			content:  []byte(XGO_GEN_IGNORE_PATTERN + " \n"),
			want:     []byte(XGO_GEN_IGNORE_PATTERN + " \n"),
			wantBool: false,
		},
		{
			name:     "pattern exists with comment",
			content:  []byte(XGO_GEN_IGNORE_PATTERN + " # comment\n"),
			want:     []byte(XGO_GEN_IGNORE_PATTERN + " # comment\n"),
			wantBool: false,
		},
		{
			name:     "content without newline",
			content:  []byte("*.log"),
			want:     []byte("*.log\n" + XGO_GEN_IGNORE_PATTERN + "\n"),
			wantBool: true,
		},
		{
			name:     "content with windows line ending",
			content:  []byte("*.log\r\n"),
			want:     []byte("*.log\r\n" + XGO_GEN_IGNORE_PATTERN + "\n"),
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := gitignoreAdd(tt.content, XGO_GEN_IGNORE_PATTERN)
			if gotBool != tt.wantBool {
				t.Errorf("gitignoreAdd() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("gitignoreAdd() got = %q, want %q", got, tt.want)
			}
		})
	}
}
