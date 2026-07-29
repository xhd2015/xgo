# mock_rule_include_as_main_module

Integration tests for `--mock-rule-include-as-main-module` /
`XGO_MOCK_RULE_INCLUDE_AS_MAIN_MODULE`.

## Layout

| Path | Role |
|------|------|
| `app/`, `lib/`, `other/` | Replace-deps (not process main) |
| `without/` | Baseline: main_module rules only → suite OK, app/other fail |
| `with_flag/` | Flag includes `app` |
| `with_env/` | Env includes `app` |
| `flag_overrides_env/` | Flag=`app`, env=`lib` → flag wins whole list |
| `with_flag_two_modules/` | Flag=`app,lib` |

Each suite uses pricing-like rules:

```text
--mock-rule {"main_module":true,"kind":"func","action":"include"}
--mock-rule {"any":true,"action":"exclude"}
```

## Run

From the xgo repo root:

```bash
go run ./script/run-test ./runtime/test/mock/mock_rule_include_as_main_module/without
go run ./script/run-test ./runtime/test/mock/mock_rule_include_as_main_module/with_flag
# …
```

## Testing note

Third-party `mock.Patch` targets go through an `interface{}` helper so static
mock-ref analysis does **not** auto-instrument them. Direct `mock.Patch(app.Hello, …)`
would instrument `app` even without include-as and would false-green these tests.
