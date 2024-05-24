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