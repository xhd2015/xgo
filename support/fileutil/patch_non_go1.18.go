//go:build !go1.18
// +build !go1.18

package fileutil

import (
	"encoding/json"
)

func PatchJSONPretty(file string, init func() interface{}, patch func(v interface{}) error) error {
	return patchJSON(file, true, init, patch)
}

func PatchJSON(file string, init func() interface{}, patch func(v interface{}) error) error {
	return patchJSON(file, true, init, patch)
}
func patchJSON(file string, pretty bool, init func() interface{}, patch func(v interface{}) error) error {
	return Patch(file, func(data []byte) ([]byte, error) {
		jsonData := init()
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
