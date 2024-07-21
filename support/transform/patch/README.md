# Syntax-aware Patch
This package provides an intuitive way to patch the go code.

The purpose is to power xgo's patching mechanism.

See [#169](https://github.com/xhd2015/xgo/issues/169#issuecomment-2241407305).

# Examples
You can find more examples in the [testdata](./testdata/) directory.

The following example is taken from [./testdata/hello_world/](./testdata/hello_world/).

`original.go`:
```go
package main

import "fmt"

func main() {
	fmt.Printf("hello world\n")
}
```

`original.go.patch`:
```go
package main

func main(){
    //append <id> fmt.Printf("the world is patched\n")
    fmt.Printf("hello world\n")
}
```

`result.go`:
```go
package main

import "fmt"

func main() {
	fmt.Printf("hello world\n")
	fmt.Printf("the world is patched\n")
}
```