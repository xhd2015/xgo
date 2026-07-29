package app

// Hello is a third-module function intended for mock.Patch under
// --mock-rule-include-as-main-module (or the matching env).
func Hello() string {
	return "app-hello"
}
