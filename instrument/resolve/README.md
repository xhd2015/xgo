# Idea
original:
```go
var SomeVar int
```

Instrumented:
```go
var SomeVar int

func SomeVar_xgo_get() int {
    __mock_res := SomeVar
    __xgo_runtime.TrapVar("SomeVar",&__mock_res)
    return __mock_res
}

func SomeVar_xgo_get_addr() *int {
    __mock_res := &SomeVar
    __xgo_runtime.TrapVarPtr("SomeVar",&__mock_res)
    return __mock_res
}
```