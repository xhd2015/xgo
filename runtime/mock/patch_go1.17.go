//go:build !go1.18
// +build !go1.18

package mock

func Patch(fn interface{}, replacer interface{}) func() {
	recvPtr, fnInfo, funcPC, trappingPC := getFunc(fn)
	return mock(recvPtr, fnInfo, funcPC, trappingPC, buildInterceptorFromPatch(recvPtr, replacer))
}
