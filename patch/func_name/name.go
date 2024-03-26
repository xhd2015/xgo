package func_name

import (
	"fmt"
)

func FormatFuncRefName(recvTypeName string, recvPtr bool, funcName string) string {
	if recvTypeName == "" {
		return funcName
	}
	if recvPtr {
		return fmt.Sprintf("(*%s).%s", recvTypeName, funcName)
	}
	return recvTypeName + "." + funcName
}
