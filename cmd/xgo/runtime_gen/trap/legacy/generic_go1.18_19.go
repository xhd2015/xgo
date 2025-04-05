//go:build go1.18 && !go1.20
// +build go1.18,!go1.20

package legacy

// for go1.18 and go1.19 only.
// generic is implemented via
// closure
const GenericImplIsClosure = true
