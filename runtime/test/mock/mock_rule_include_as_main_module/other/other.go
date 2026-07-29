package other

// Hello is a dependency that is never listed in include-as-main-module.
// With main_module-only include rules, mock.Patch must keep failing.
func Hello() string {
	return "other-hello"
}
