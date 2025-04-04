package ast

import (
	"strconv"
	"strings"
)

func QuoteNames(names []string) []string {
	quotedNames := make([]string, len(names))
	for i, name := range names {
		quotedNames[i] = strconv.Quote(name)
	}
	return quotedNames
}

func JoinQuoteNames(names []string, sep string) string {
	return strings.Join(QuoteNames(names), sep)
}
