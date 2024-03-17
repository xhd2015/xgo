//go:build go1.18
// +build go1.18

package fileutil

import (
	"encoding/json"
)

func PatchJSONPretty[T any](file string, patch func(v *T) error) error {
	return patchJSON(file, true, patch)
}

func PatchJSON[T any](file string, patch func(v *T) error) error {
	return patchJSON(file, true, patch)
}
func patchJSON[T any](file string, pretty bool, patch func(v *T) error) error {
	return Patch(file, func(data []byte) ([]byte, error) {
		var jsonData T
		if len(data) > 0 {
			err := json.Unmarshal(data, &jsonData)
			if err != nil {
				return nil, err
			}
		}
		err := patch(&jsonData)
		if err != nil {
			return nil, err
		}
		if pretty {
			return json.MarshalIndent(jsonData, "", "    ")
		}
		return json.Marshal(jsonData)
	})
}
