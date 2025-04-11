package template

import (
	"github.com/xhd2015/xgo/support/goinfo"
)

func AppendGetFuncNameImpl(goVersion *goinfo.GoVersion, content string) string {
	// func name patch
	if goVersion.Major > goinfo.GO_MAJOR_1 || goVersion.Minor > goinfo.GO_VERSION_24 {
		panic("should check the implementation of runtime.FuncForPC(pc).Name() to ensure __xgo_get_pc_name is not wrapped in print format above go1.23,it is confirmed that in go1.21~go1.24 the name is wrapped in funcNameForPrint(...).")
	}
	if goVersion.Major > 1 || goVersion.Minor >= 21 {
		content += RuntimeGetFuncName_Go121
	} else if goVersion.Major == 1 {
		if goVersion.Minor >= 17 {
			// go1.17,go1.18,go1.19
			content += RuntimeGetFuncName_Go117_120
		}
	}
	return content
}
