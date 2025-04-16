package instrument_var

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/constants"
)

func genCode(fileIndex int, varPrefix string, varName string, infoVar string, varType string) string {
	var lines = []string{
		`func %s%s_xgo_get() %s {`,
		`__mock_res := %s`, ";",
		constants.LINK_TRAP_VAR + `%d(%s,&%s,&__mock_res)`, ";",
		`return __mock_res`, ";",
		`}`, ";",
		`func %s%s_xgo_get_addr() *%s {`,
		`__mock_res := &%s`, ";",
		constants.LINK_TRAP_VAR_PTR + `%d(%s,&%s,&__mock_res)`, ";",
		`return __mock_res`, ";",
		`}`,
	}
	template := strings.Join(lines, "")
	return fmt.Sprintf(template,
		varPrefix, varName, varType,
		varName,
		fileIndex, infoVar, varName,
		varPrefix, varName, varType,
		varName,
		fileIndex, infoVar, varName,
	)
}
