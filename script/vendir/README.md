# Vendir
vendir is a utility helps to introduce third party vendor libraries without introducing dependencies in the go.mod.

This way, a library itself can separate its exported APIs from its internal dependencies.

# Usage
```sh
go run github.com/xhd2015/xgo/script/vendir@latest create ./internal/third/src ./internal/third/vendir
```
This command will copy all dependencies in `./internal/third/src/vendor` to `./internal/third/vendir`, and rewrite all imports to that created internal package.

Prior to running this command, you should create a go.mod, add dependency and run `go mod vendor` inside `./internal/third/src`.

# How it works?
In general, the source directory should contain a go.mod and a vendor directory describing all it's dependencies.

The target directory then will be created by copying all dependencies from source vendor directory and rewrite all import paths(except stdlib) to internal paths.

This results in all third party code self-included, so go.mod does not change at all.

Check [./example](./example) for demonstration.

# Why?
The xgo project itself provides a range of utilities, among which the incremental coverage tool depends on heavily a lot of external APIs. 

But we don't want to expose these dependencies to the end user. Initially we took an approach that compiles the incremental coverage tool standalone and download it when required. But that has obvious shortcomings, such as bad maintenance and underperformed user experience.

We take the step further, why binray dependency? Isn't there a way to statically compile all code together?

So we came up with this vendir utility.