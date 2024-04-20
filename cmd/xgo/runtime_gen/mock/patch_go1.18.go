//go:build go1.18
// +build go1.18

package mock

// TODO: what if `fn` is a Type function
// instead of an instance method?
// func Patch[T any](fn T, replacer T) func() {
// 	recvPtr, fnInfo, funcPC, trappingPC := getFunc(fn)
// 	return mock(recvPtr, fnInfo, funcPC, trappingPC, buildInterceptorFromPatch(recvPtr, replacer))
// }

// NOTE: as a library targeting under go1.18, the library itself should not
// use any generic thing
//
// situiation:
//    go.mod: 1.16
//    runtime/go.mod: 1.18
//    go version: 1.20

// compile error:
//  implicit function instantiation requires go1.18 or later (-lang was set to go1.16; check go.mod)
//     mock.Patch(...)
//   because mock.Patch was defined as generic
