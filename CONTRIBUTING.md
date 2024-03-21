# Contributing to xgo
Thanks for helping, this document helps you get started. 

# Development guide
First, clone this repository:
```sh
git clone https://github.com/xhd2015/xgo
cd cgo
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
go run ./cmd/xgo help

# run Hello world test
go test -run TestHelloWorld -v ./test
```

If you want to check instrumented GOROOT, run:
```sh
go run ./script/setup-dev
```

The above command will prepare a instrumented GOROOT and print the directory. 

You can open that directory and check the internals.