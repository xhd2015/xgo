package xgo

import "embed"

// since go:embed does not work with ../ paths
// so we add a global asset here.
// https://github.com/golang/go/issues/46056

//go:embed patch
var PatchFS embed.FS
