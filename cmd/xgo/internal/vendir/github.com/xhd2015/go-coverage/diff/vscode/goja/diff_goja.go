package goja

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/dop251/goja"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/diff/vscode"
)

var gojaProgram = goja.MustCompile("diff_goja.js", vscode.DiffGojaCode, true)

func UseGojaDiff() {
	vscode.DiffImpl = Diff
}

func Diff(req *vscode.Request) (*vscode.Result, error) {
	runtime := goja.New()
	err := runtime.Set("request", req)
	if err != nil {
		return nil, err
	}
	res, err := runtime.RunProgram(gojaProgram)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("res: %T, %+v", res, res)
	var goRes vscode.Result
	err = runtime.ExportTo(res, &goRes)
	if err != nil {
		return nil, err
	}

	return &goRes, nil
}
