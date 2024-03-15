# Function Explained
Guess what's the output of the following code:
```go
package main

import (
    "fmt"
    "unsafe"
    "runtime"
)

func main(){
    ipc := (**uintptr)(unsafe.Pointer(&example))
    pc := *ipc
    fnName := runtime.FuncForPC(*pc).Name()
    fmt.Printf
}

func example(){
}
```

The name `example` refers to a global variable whose type is `func()`, and value is a pointer to its entrypoint,i.e. `pc`.

Also, there is a nicer way to write the code:
```go
fn := runtime.FuncForPC(reflect.ValueOf(example).Pointer()).Name()
```

`reflect.Value.Pointer()`: returns the value as pointer.

Actually, the `reflect.ValueOf(fn)` is what reveals that the func. Note that `reflect.ValueOf()` takes `interface{}` as an argument, so first we need to convert `example` to an `interface{}`.

An `interface{}` consists of two fields, which can be represented as:
```go
type intf struct {
		typ  uintptr
		pc *uintptr
}
```

(TO BE FINISHED)