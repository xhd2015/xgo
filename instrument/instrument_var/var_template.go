package instrument_var

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/constants"
)

func genCode(varPrefix string, varName string, infoVar string, varType string) string {
	var lines = []string{
		`func %s%s_xgo_get() %s {`,
		`__mock_res := %s`, ";",
		constants.RUNTIME_PKG_NAME_VAR + `.XgoTrapVar(` + constants.UNSAFE_PKG_NAME_VAR + `.Pointer(%s),&%s,&__mock_res)`, ";",
		`return __mock_res`, ";",
		`}`, ";",
		`func %s%s_xgo_get_addr() *%s {`,
		`__mock_res := &%s`, ";",
		constants.RUNTIME_PKG_NAME_VAR + `.XgoTrapVarPtr(` + constants.UNSAFE_PKG_NAME_VAR + `.Pointer(%s),&%s,&__mock_res)`, ";",
		`return __mock_res`, ";",
		`}`,
	}
	template := strings.Join(lines, "")
	return fmt.Sprintf(template,
		varPrefix, varName, varType,
		varName,
		infoVar, varName,
		varPrefix, varName, varType,
		varName,
		infoVar, varName,
	)
}
