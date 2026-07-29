package lib

// Hello is a second replace dependency used for multi-module include lists
// and flag-vs-env override cases.
func Hello() string {
	return "lib-hello"
}
