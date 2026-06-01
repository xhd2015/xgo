# Add Patches for New Go Version

The next go version at https://go.dev/dl/ has already been released(check with ` go run ./script/download-go download list`),  a new corresponding patch for this go version under patches/ directory needs to be added.

# Seed Patch and Prepare Source Code

Please first run `cp -R ./patches/<current-latest-go-version> ./patches/<next-go-version>` to seed initial patches for the new go version.

Then check ./go-release to see if the source code of the two go versions are available, if not, download them, e.g. `go run ./download-go go1.25.1`.

# Research and Adjust the Patches

Read PATCH_DSL.md to understand the patch DSL.

Compare the source code changes from current latest go version to next go version, and understand the semantic changes in the source code.

Then check how the patches needs to be adjusted to reflect the go source code changes, and ask for user approval.

Once approved, proceed to implementation.

# Validate

After adjusted the patches, validate with:

```sh
# run all tests
go run ./script/run-test --with-goroot ./go-release/<next-go-version>
```

If this step failed, go back to last step and retry;

If this step is ok, proceed to next step.

# Regression Validation

To avoid regression, run final test

```sh
# NOTE these two integration tests only available for go1.24 and go1.25, they're run as regression baseline
go run ./test/integrations/test-file-patch-generated-same-diffs-as-programatic-patch/ --go-version 1.24
go run ./test/integrations/test-file-patch-can-be-repeated-on-patched-goroot --go-version 1.24
```

# Summarize to CHANGELOG

After finished code changes, conclude to `./patches/<next-go-version>/CHANGELOG` how the patches were adjusted to reflect source code changes.