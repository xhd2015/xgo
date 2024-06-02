package cmd

import "strings"

// quote for shell print, don't use it
// to quote args passed to go cmd
// NOTE: this is for bash-like shells,
// not tested against PowerShell and CMD on windows
func Quote(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, "\t \n;<>\\${}()&!*") { // special args
		s = strings.ReplaceAll(s, "'", "'\\''")
		return "'" + s + "'"
	}
	return s
}
