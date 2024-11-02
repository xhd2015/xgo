package ast

import "strings"

func BuildArgsToSyntaxArgs(args []string) []string {
	if len(args) == 0 {
		return []string{"./..."}
	}
	newArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasPrefix(arg, "./") || strings.HasSuffix(arg, "...") {
			newArgs = append(newArgs, arg)
			continue
		}
		newArgs = append(newArgs, strings.TrimSuffix(arg, "/")+"/...")
	}
	return newArgs
}
