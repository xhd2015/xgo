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
    "args":["--my-program-env","TEST"],
    "mock_rules":[{
        "stdlib": true,
        "action": "exclude"
    }],
    "xgo":{
        "auto_update": true
    },
    "coverage": {
        "diff_with": "origin/master"
    }
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

## `bypass_go_flags`
When running test, this will add a leading `--` before the `args` list, effectively bypassing go's builtin flag parsing procedure.

The reason why this is needed, is that a go built test binary will parse every flag on the command line, which terminates the program upon missing a flag, with complaining about it. However, that flag is expected to be parsed by the program itself.

See https://github.com/xhd2015/xgo/issues/263.

Default: false.

## `mock_rules`
A list of `Rule` config to specify which packages and functions can be mocked.

The `Rule` definition:
```json
{
    "pkg": "",
    "func": "",
    "kind": "func" | "var" | "const",
    "stdlib": true | false,
    "main_module": true | false,
    "action": "" || "include" | "exclude" 
}
```

A practical example to only mock functions of main module and some RPC functions:
```json
{
    "mock_rules": [
        {
            "main_module": true,
            "kind": "func",
            "action": "include"
        },
        {
            "main_module": true,
            "kind": "var,const",
            "pkg": "github.com/xhd2015/demo/config//**",
            "action": "include"
        },
        {
            "comment": "protobuf rpc",
            "pkg": "github.com/xhd2015/external/protobuf3.pb/**",
            "action": "include"
        },
        {
            "comment": "log",
            "pkg": "github.com/xhd2015/external/framework_logger/impl",
            "action": "include"
        },
        {
            "any": true,
            "action": "exclude"
        }
    ]
}
```

Default: `null`

## `xgo`
Configuration of xgo behavior.

Definition:
```json
{
    "auto_update": true | false(default)
}
```

## `coverage`
Definition:
```json
{
    "disabled": true
}
```

By default, if `coverage` is missing or `null`, then coverage is enabled.

Setting `"coverage": false` or `"coverage":{"disabled": true}` will disable it.