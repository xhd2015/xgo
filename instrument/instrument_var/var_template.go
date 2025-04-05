package instrument_var

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/constants"
)

func genCode(varName string, infoVar string, varType string) string {
	var lines = []string{
		`func %s_xgo_get() %s {`,
		`__mock_res := %s`, ";",
		constants.RUNTIME_PKG_NAME_VAR + `.XgoTrapVar(` + constants.UNSAFE_PKG_NAME_VAR + `.Pointer(%s),&%s,&__mock_res)`, ";",
		`return __mock_res`, ";",
		`}`, ";",
		`func %s_xgo_get_addr() *%s {`,
		`__mock_res := &%s`, ";",
		constants.RUNTIME_PKG_NAME_VAR + `.XgoTrapVarPtr(` + constants.UNSAFE_PKG_NAME_VAR + `.Pointer(%s),&%s,&__mock_res)`, ";",
		`return __mock_res`, ";",
		`}`,
	}
	template := strings.Join(lines, "")
	return fmt.Sprintf(template,
		varName, varType,
		varName,
		infoVar, varName,
		varName, varType,
		varName,
		infoVar, varName,
	)
}
