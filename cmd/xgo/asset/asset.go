package asset

import "embed"

// since go:embed does not work with ../ paths
// so we add a global asset here.
// https://github.com/golang/go/issues/46056
//
// initially we put an asset.go under the root module
// but feels too silly, the root should be clean for
// a serious foundational tool, so use generate to
// selectively copy into this dir

const (
	CompilerPatchGen = "compiler_patch_gen"
	RuntimeGen       = "runtime_gen"
	Patches          = "patches"
)

//go:embed compiler_patch_gen
var CompilerPatchGenFS embed.FS

//go:embed runtime_gen
var RuntimeGenFS embed.FS

//go:embed patches
var PatchesFS embed.FS
