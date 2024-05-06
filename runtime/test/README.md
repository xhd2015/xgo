# Tests of the runtime package
Runnable tests are listed in file [../../script/run-test/main.go](../../script/run-test/main.go).

# Run tests
```sh
# all

# specific test
go run -tags dev ./cmd/xgo test --with-goroot go1.17.13 --project-dir ./runtime/test/mock_var -v -run TestThirdPartyTypeMethodVar
```