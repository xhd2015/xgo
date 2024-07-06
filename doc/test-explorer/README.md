# Test Explorer
Test Explorer is a powerful UI tool to execute go test in a tree-view.

It helps debug go test more easily.

# `test.config.json`
When executing test from Test Explorer, xgo will read configuration from `test.config.json` found from the project root(alongside with `go.mod`) if any.

An example config:
```json
{
    "go":{
        "min":"1.18",
        "max":"1.20"
    },
    "go_cmd":"xgo",
    "exclude": [
        "integration_test",
    ],
    "env":{
        "TEST_WITH_XGO":"true"
    },
    "flags":["-p=12"],
    "args":["--my-program-env","TEST"]
}
```

Descriptions for each fields:

## `go.min`, `go.max`
The minimal and maximal go version that can be used to execute project's tests.

If current go version does not satisfy the requirement, the test explorer will print an error and exits.

Default: `""`.

## `go_cmd`
By default, xgo will uses `xgo test` to execute the project's tests. 

If one doesn't want to use Mock functionalities provided by `xgo`, instead only want to use the test explorer itself, then set `"go_cmd": "go"`, the original `go test` will be used.

Default: `"xgo"`.

## `exclude`
A list of patterns to be ignored when listing test cases in the test explorer UI.

For example, if there is an `integration_test` directory which contains heavy integration test, and it should be excluded when executing unit test.

Then setting `"exclude":["integration_test"]` will hide tests from the `integration_test` directory and it's children.

Default: `null`.

## `flags`
A list of flags that will be passed to `xgo test`.

For example, `-failfast` to fail quickly if any test fails. 

Use `go help test` and `go help testflag` to get a full list of potential flags.

Default: `null`.

## `args`
A list of args that will be passed to the test binary itself.

This option is a symptom to `go test -args <args>...`.

Default: `null`.