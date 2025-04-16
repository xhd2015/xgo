# Cheatsheet
Prepare:
```sh
rm -rf ~/GOROOT_DEBUG_24
cp -R $X/xgo/go-release/go1.24.1 ~/GOROOT_DEBUG_24
go install ./script/xgo.helper
xgo.helper setup-vscode ~/GOROOT_DEBUG_24

cd runtime/test/build/overlay_build_cache_error_with_go
```



Run first to verify the problem exist:
```sh
# run first without overlay
xgo.helper run-go ~/GOROOT_DEBUG_24 test -v ./overlay_test_with_gomod

# then overlay
xgo.helper run-go ~/GOROOT_DEBUG_24 test -v -overlay ./overlay_gomod_first.json ./overlay_test_with_gomod
```

Second run expect error output:
```sh
overlay_build_cache_error_with_go/overlay/reverse/reverse.go:6:24: could not import runtime (open : no such file or directory)
```

Debug:
```sh
GOMODCACHE=$PWD/.xgo/gen/gomodcache xgo.helper debug-compile ~/GOROOT_DEBUG_24 golang.org/x/example/hello/reverse test -v -overlay ./overlay_gomod_first.json -gcflags="golang.org/x/example/hello/reverse=-v" ./overlay_test_with_gomod
```