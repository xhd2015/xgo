package asset

import "embed"

// since go:embed does not work with ../ paths
// so we add a global asset here.
// https://github.com/golang/go/issues/46056

const (
	CompilerPatchGen = "compiler_patch_gen"
	RuntimeGen       = "runtime_gen"
)

//go:embed compiler_patch_gen
var CompilerPatchGenFS embed.FS

//go:embed runtime_gen
var RuntimeGenFS embed.FS
