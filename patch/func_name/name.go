package func_name

import (
	"fmt"
	"strings"
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

// a/b/c.A
// a/b/c.(*C).X
// a/b/c.C.Y
// a/b/c.Z
func ParseFuncName(fullName string, hasPkg bool) (pkgPath string, recvName string, recvPtr bool, funcName string) {
	s := fullName
	funcNameDot := strings.LastIndex(s, ".")
	if funcNameDot < 0 {
		funcName = s
		return
	}
	funcName = s[funcNameDot+1:]
	s = s[:funcNameDot]

	recvName = s
	if hasPkg {
		recvDot := strings.LastIndex(s, ".")
		if recvDot < 0 {
			pkgPath = s
			return
		}
		recvName = s[recvDot+1:]
		s = s[:recvDot]
	}

	// when the recvName is from system readonly area,
	// these two line panics in debug mode, don't know why
	//
	// recvName = strings.TrimPrefix(recvName, "(")
	// recvName = strings.TrimSuffix(recvName, ")")
	if len(recvName) > 0 && recvName[0] == '(' {
		recvName = recvName[1:]
	}

	if len(recvName) > 0 && recvName[len(recvName)-1] == ')' {
		recvName = recvName[:len(recvName)-1]
	}

	if strings.HasPrefix(recvName, "*") {
		recvPtr = true
		recvName = recvName[1:]
	}
	pkgPath = s

	return
}
