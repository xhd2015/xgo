//go:build go1.18
// +build go1.18

package mock

// TODO: what if `fn` is a Type function
// instead of an instance method?
func Patch[T any](fn T, replacer T) func() {
	return Mock(fn, buildInterceptorFromPatch(replacer))
}
