# About

This package ports the vscode-diff(https://github.com/microsoft/vscode/tree/main/src/vs/editor/common/diff) into go, so that we can get a consistent diff view in backend and in frontend which uses the monaco-editor, the web version of vscode.

# Use as cmd

```bash
# run go generate ./... first
$ go generate ./...

# use gen/diff.js as a parser
$ echo -ne '{"oldLines":["A","B","C"],"newLines":["A","B1","C"]}' | node diff/vscode/gen/diff.js
{"quitEarly":false,"changes":[{"originalStartLineNumber":2,"originalEndLineNumber":2,"modifiedStartLineNumber":2,"modifiedEndLineNumber":2}]}

# diff_v2
$ echo -ne '{"oldLines":["A","B","C"],"newLines":["A","B1","C"]}' | node diff/vscode/gen/diff_v2.js
{"quitEarly":false,"changes":[{"originalStartLineNumber":2,"originalEndLineNumber":2,"modifiedStartLineNumber":2,"modifiedEndLineNumber":2}]}
```

# Use goja

The goja implementation comes with [goja](https://github.com/dop251/goja), a javascript runtime implemented in go, but needs go1.16.

If you have go1.16 you can opt in with that:

```go
...
import (
    "github.com/xhd2015/go-coverage/diff/vscode/goja"
)
func init(){
	goja.UseGojaDiff()
}
```

The performance is in the middle:

```
native go: 1585 ns/op = 0.001ms/op
goja: 590299 ns/op = 0.59ms/op
the stdin-stdout:  109311564 ns/op = 109ms/op
```

# Use as a standby servant of golang

```go
package main

import (
   vscode_diff "gitub.com/xhd2015/go-coverage/diff/vscode"
)

func main(){
    defer vscode_diff.Destroy()
	res, err := vscode_diff.Diff(&Request{
		OldLines: []string{"A", "B", "C"},
		NewLines: []string{"A", "B1", "C"},
	})
	if err != nil {
		panic(err)
	}
	resJSONBytes, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	resJSON := string(resJSONBytes)
	resJSONExpect := `{"quitEarly":false,"changes":[{"originalStartLineNumber":2,"originalEndLineNumber":2,"modifiedStartLineNumber":2,"modifiedEndLineNumber":2}]}`
	if resJSON != resJSONExpect {
		log.Fatalf("expect %s = %+v, actual:%+v", `resJSON`, resJSONExpect, resJSON)
	}
}
```

# Performance

There are 2 benchmarks for myers diff and vscode diff, the result is as following:

```
myers   (go native):                   1585 ns/op
vscode  (nodejs implementation, ipc):  10845941 ns/op = 10.8ms/op
```

The benchmark is showing that using vscode-diff in go with ipc to nodejs is 6842x slower than that of native go implementation.

Since diff is not a very frequent operation and usually is computed in background, so this latency is acceptable.

For realtime scenerios, should take this downgrade performance into consideraion.

# NOTE

Current implementation uses terminal io, and it will start a background nodejs process to listen for requests.
So it depends on `node`, and there may exist process leaking, meaning when the main process exits, unless explicitly
calling `vscode_diff.Destroy()`, the background may keep alive.

But since this package is meant to be used in cloud server so it is not a big problem.
Anyone who wants to use this in regular way should add a `defer vscode_diff.Destroy()` in the `main` function
to avoid possible process leak.

Update in 2022-12-20: Now there is a ping mechanism that when parent go process spawn the nodejs process, it will periodically
sends ping message to that child. If the child fails to detect ping in 10s, it will ends itself.
The parent go process will ping at every 5s.

```bash
# dev: 177K
webpack --config server/webpack.config-difflib.js --progress --mode=production

# min: 73K
webpack --config server/webpack.config-difflib.js --progress --mode=development --devtool source-map
```
