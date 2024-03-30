//go:build go1.17 && !go1.18
// +build go1.17,!go1.18

package trap

// for go1.17, a method
// is wrapped in two layer:
//
//	X-fm:
//	   trapped
//	   call X
//	X:
//	   trapped
//
// while in go1.18 and above, only X is trapped, X-fm is not trapped
const methodHasBeenTrapped = true
