package main

import (
	"reflect"
	"testing"
)

func TestAppendShortIfFast(t *testing.T) {
	tests := []struct {
		name       string
		remainArgs []string
		want       []string
	}{
		{name: "no existing -short", remainArgs: []string{"-v"}, want: []string{"-v", "-short"}},
		{name: "with existing -short", remainArgs: []string{"-v", "-short"}, want: []string{"-v", "-short"}},
		{name: "empty remain args", remainArgs: nil, want: []string{"-short"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendShortIfFast(tt.remainArgs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestParseUseFilePatchesFlag(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		want    *bool
		wantErr bool
	}{
		{name: "true", val: "true", want: ptrBool(true)},
		{name: "false", val: "false", want: ptrBool(false)},
		{name: "empty (bare flag)", val: "", want: ptrBool(true)},
		{name: "invalid", val: "bad", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUseFilePatchesFlag(tt.val)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if *got != *tt.want {
				t.Fatalf("expected %v, got %v", *tt.want, *got)
			}
		})
	}
}

