package ctxt

func PkgWithinModule(pkgPath string, module string) (suffix string, ok bool) {
	n := len(pkgPath)
	m := len(module)
	if n < m {
		return "", false
	}
	// check prefix
	for i := 0; i < m; i++ {
		if pkgPath[i] != module[i] {
			return "", false
		}
	}
	if n == m {
		return "", true
	}
	if pkgPath[m] != '/' {
		return "", false
	}
	return pkgPath[m+1:], true
}
