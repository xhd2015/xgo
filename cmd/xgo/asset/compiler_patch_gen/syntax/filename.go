package syntax

import "cmd/compile/internal/syntax"

var AbsFilename func(name string) string        // only set under go1.18, will be nil for >= go1.18
var TrimFilename func(b *syntax.PosBase) string // will be set >= go1.18
