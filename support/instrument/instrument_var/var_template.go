package instrument_var

import (
	"fmt"
	"strings"
)

func genCode(varName string, varType string) string {
	var lines = []string{
		`func %s_xgo_get() %s {`,
		`__mock_res := %s`, ";",
		`__xgo_var_runtime.XgoTrapVar(%q,&%s,&__mock_res)`, ";",
		`return __mock_res`, ";",
		`}`, ";",
		`func %s_xgo_get_addr() *%s {`,
		`__mock_res := &%s`, ";",
		`__xgo_var_runtime.XgoTrapVarPtr(%q,&%s,&__mock_res)`, ";",
		`return __mock_res`, ";",
		`}`,
	}
	template := strings.Join(lines, "")
	return fmt.Sprintf(template,
		varName, varType,
		varName,
		varName, varName,
		varName, varType,
		varName,
		varName, varName,
	)
}
