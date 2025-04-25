package syntax

import (
	"cmd/compile/internal/xgo_rewrite_internal/patch/info"
)

// split fileDecls to a list of batch
// when statements gets large, it will
// exceeds the compiler's threshold, causing
//
//	internal compiler error: NewBulk too big
//
// see https://github.com/golang/go/issues/33437
// see also: https://github.com/golang/go/issues/57832 The input code is just too big for the compiler to handle.
// here we split the files per 1000 functions
func splitBatch(funcDecls []*info.DeclInfo, batch int) [][]*info.DeclInfo {
	if batch <= 0 {
		panic("invalid batch")
	}
	n := len(funcDecls)
	if n <= batch {
		// fast path
		return [][]*info.DeclInfo{funcDecls}
	}
	var res [][]*info.DeclInfo

	var cur []*info.DeclInfo
	for i := 0; i < n; i++ {
		cur = append(cur, funcDecls[i])
		if len(cur) >= batch {
			res = append(res, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		res = append(res, cur)
		cur = nil
	}
	return res
}
