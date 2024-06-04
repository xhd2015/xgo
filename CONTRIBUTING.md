# Contributing to xgo
Thanks for helping, this document helps you get started. 

# Development guide
First, clone this repository:
```sh
git clone https://github.com/xhd2015/xgo
cd xgo
```

Then, setup git hooks:
```sh
go run ./script/git-hooks install

chmod +x .git/hooks/pre-commit
chmod +x .git/hooks/post-commit
```

All are set, now start to development, try:
```sh
# help
go run -tags dev ./cmd/xgo help

# run Hello world test
go test -tags dev -run TestHelloWorld -v ./test
```

NOTE: when developing, always add `-tags dev` to tell go that we are building in dev mode.

If you want to check instrumented GOROOT, run:
```sh
go run ./script/setup-dev
```

The above command will prepare a instrumented GOROOT and print the directory. 

You can open that directory and check the internals.

# Adding Tests Before Adding Feature
Xgo prefers TDD to bring new features. 

We suggest every feature to be tested exhaustively.

To run all tests of the xgo project:
```sh
go run ./script/run-test
```

This will run all tests with all go versions found at the directory `go-release`.

We can also explicitly specify all expected go versions we want to pass:
```sh
go run ./script/run-test/ --include go1.17.13 --include go1.18.10 --include go1.19.13 --include go1.20.14 --include go1.21.8 --include go1.22.1
```

If there were testing cache, we can force the test to re-run by adding a `-count=1` flag:
```sh
go run ./script/run-test/ --include go1.17.13 --include go1.18.10 --include go1.19.13 --include go1.20.14 --include go1.21.8 --include go1.22.1 -count=1
```

If a go version is not found in `go-release`, we can download it with:
```sh
go run ./script/download-go go1.22.1
```

# Install xgo from source
Just clone the repository, and run:
```sh
go install ./cmd/xgo
```

It's totally the same as `go install github.com/xhd2015/xgo/cmd/xgo@latest`, but for local.

# Debug the go compiler
First, build a package with `--debug-compile` flag:
```sh
go run -tags dev ./cmd/xgo test -c --debug-compile --project-dir runtime/test/debug
```

Then, run `go-tool-debug-compile`
```sh
go run ./cmd/go-tool-debug-compile
```


Output:
```log
dlv listen on localhost:2345
Debug with IDEs:
  > VSCode: add the following config to .vscode/launch.json configurations:
    {
        "configurations": [
                {
                        "name": "Debug dlv localhost:2345",
                        "type": "go",
                        "debugAdapter": "dlv-dap",
                        "request": "attach",
                        "mode": "remote",
                        "port": 2345,
                        "host": "127.0.0.1",
                        "cwd":"./"
                }
        }
    }
    NOTE: VSCode will map source files to workspace's goroot, which causes problem when debugging go compiler.
      To fix this, update go.goroot in .vscode/settings.json to:
       /Users/xhd2015/.xgo/go-instrument-dev/go1.21.7_Us_xh_in_go_096be049/go1.21.7
      And set a breakpoint at:
       /Users/xhd2015/.xgo/go-instrument-dev/go1.21.7_Us_xh_in_go_096be049/go1.21.7/src/cmd/compile/main.go
  > GoLand: click Add Configuration > Go Remote > localhost:2345
  > Terminal: dlv connect localhost:2345
```

Following these instructions, using your favorite IDE like VSCode,GoLand or just terminal to debug:
<img width="1792" alt="image" src="https://github.com/xhd2015/xgo/assets/14964938/673df393-6632-4eed-a004-400e0c70d0d1">

