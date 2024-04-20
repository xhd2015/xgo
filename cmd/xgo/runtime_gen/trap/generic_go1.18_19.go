//go:build go1.18 && !go1.20
// +build go1.18,!go1.20

package trap

// for go1.18 and go1.19 only.
// geneirc is implemented via
// closure
const GenericImplIsClosure = true
