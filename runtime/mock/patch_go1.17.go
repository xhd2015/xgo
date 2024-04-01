//go:build !go1.18
// +build !go1.18

package mock

func Patch(fn interface{}, replacer interface{}) func() {
	return Mock(fn, buildInterceptorFromPatch(replacer))
}
