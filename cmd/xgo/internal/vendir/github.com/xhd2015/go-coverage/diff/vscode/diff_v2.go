package vscode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sync"
)

var initV2Once sync.Once
var diffV2File string
var initErr error

// user can set another DiffImpl
var DiffImpl = DiffV1

func Diff(req *Request) (*Result, error) {
	// return DiffV2(req)
	return DiffImpl(req)
}

func DiffV2(req *Request) (*Result, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(reqJSON)
	res, err := runSyncCmd(reader)
	if err != nil {
		return nil, err
	}
	var r Result
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func runSyncCmd(input io.Reader) (output []byte, err error) {
	initV2Once.Do(func() {
		diffV2File, initErr = InitCode("diff_v2.js", diffV2JSCode)
	})
	if initErr != nil {
		return nil, initErr
	}
	cmd = exec.Command("node", diffV2File)
	cmd.Stdin = input
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
func InitCode(file string, code string) (jsFile string, err error) {
	tmpFile, err := os.MkdirTemp(os.TempDir(), "vscode-diff")
	if err != nil {
		return "", fmt.Errorf("cannot create vscode-diff dir: %v", err)
	}
	diffFile = path.Join(tmpFile, file)
	err = ioutil.WriteFile(diffFile, []byte(code), 0777)
	if err != nil {
		return "", fmt.Errorf("cannot create file:%s %v", diffFile, err)
	}
	return diffFile, nil
}
