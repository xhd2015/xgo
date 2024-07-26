package assert

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
)

// Diff compares two values, resulting in a human readable diff format.
func Diff(expected interface{}, actual interface{}) string {
	if res, ok := tryDiffError(expected, actual); ok {
		return res
	}
	// check if any is error type
	expectedTxt, err := toDiffableText(expected)
	if err != nil {
		return err.Error()
	}
	actualTxt, err := toDiffableText(actual)
	if err != nil {
		return err.Error()
	}
	diff, err := diffText([]byte(expectedTxt), []byte(actualTxt))
	if err != nil {
		sep := ""
		if diff != "" {
			sep = ", "
		}
		diff += sep + "err: " + err.Error()
	}
	return diff
}

func toDiffableText(v interface{}) (string, error) {
	if v == nil {
		return "nil", nil
	}
	if v, ok := toPlainString(v); ok {
		return tryPrettyJSONText(v)
	}
	text, err := prettyObj(v)
	if err != nil {
		return "", fmt.Errorf("marshal %T: %v", v, err)
	}
	return string(text), nil
}

func prettyObj(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func toPlainString(v interface{}) (string, bool) {
	switch v := v.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case json.RawMessage:
		return string(v), true
	}

	// slow
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.String(), true
	case reflect.Slice:
		if rv.IsNil() {
			return "null", true
		}

		if rv.Type().Elem().Kind() == reflect.Uint8 {
			// []byte
			return string(rv.Bytes()), true
		}
	}
	return "", false
}

func tryPrettyJSONText(s string) (string, error) {
	if s == "" || s == "null" {
		return s, nil
	}
	var v interface{}
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		return s, nil
	}
	text, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func diffText(expected []byte, actual []byte) (string, error) {
	tmpDir, err := os.MkdirTemp("", "diff")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	err = fileutil.WriteFile(filepath.Join(tmpDir, "expected"), expected)
	if err != nil {
		return "", err
	}
	err = fileutil.WriteFile(filepath.Join(tmpDir, "actual"), actual)
	if err != nil {
		return "", err
	}

	diff, err := cmd.Dir(tmpDir).Output("git", "diff", "--no-index", "--no-color", "--", "expected", "actual")
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			err = nil
		}
	}
	return diff, err
}
