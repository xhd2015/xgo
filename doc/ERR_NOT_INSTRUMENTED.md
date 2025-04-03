# Cause of `error not instrumented by xgo`
When mock setup fails:
```go
package main

import "third.party/pkg"

func TestThirdPackage(t *testing.T){
    mock.Patch(pkg.DoSomething, func() string{
        return "mock"
    })
}
```

With the following msg:
```

```

It means xgo does not get

# How to solve?
By default xgo will only insert trap for packages of main module, which is resolved by `go list ./...`.

Packages not trapped by xgo cannot be mocked, the setup naturally fails.

To add extra packages for trapping, you can specify `--trap pkg`:
```sh
xgo test --trap third.party/pkg <remaining args...>
```

To add more packages, repeat this option multiple times.