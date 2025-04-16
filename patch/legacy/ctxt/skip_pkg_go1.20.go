//go:build go1.20
// +build go1.20

package ctxt

func isSkippableSpecialPkg(pkgPath string) bool {
	return false
}
