package fileutil

import "os"

func UpdateFile(path string, updateFunc func(content []byte) (bool, []byte, error)) error {
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		if !os.IsNotExist(readErr) {
			return readErr
		}
	}
	ok, updated, err := updateFunc(content)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	return os.WriteFile(path, updated, 0644)
}
