# Test using xgo 
The xgo_test directory are for testing if xgo link names correctly.

To test:
```sh
# run all
go run ./cmd/xgo test ./test/xgo_test/...

# run specific one
go run ./cmd/xgo test -run TestMethodValueCompare -v ./test/xgo_test/...
```

Or use the run-test script:
```sh
go run ./script/run-test/ --xgo-test-only -count=1
```


To debug the test:
```sh
go run ./cmd/xgo test -c -o debug.bin ./test/xgo_test/method_value_cmp

dlv exec ./debug.bin -test.v
```