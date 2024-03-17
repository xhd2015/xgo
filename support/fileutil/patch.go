package fileutil

import (
	"bytes"
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
