package fileutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

func Patch(file string, patch func(data []byte) ([]byte, error)) error {
	data, dataErr := ioutil.ReadFile(file)
	if dataErr != nil {
		if !errors.Is(dataErr, os.ErrNotExist) {
			return dataErr
		}
	}
	newData, err := patch(data)
	if err != nil {
		return err
	}
	if bytes.Equal(newData, data) {
		return nil
	}
	return ioutil.WriteFile(file, []byte(newData), 0755)
}

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
