//go:build !go1.24

package asset

import "embed"

// PatchesFS is empty on Go < 1.24, because file-based patch
// was introduced in go1.25 support, and back ported to go1.24
// and for <= go1.17, "all:" prefix in go:embed is not supported; the "all:" prefix in go:embed
// requires Go 1.18+,
// leave an empty nil FS is fine — a runtime error will occur if someone
// attempts --use-file-patches with a pre-1.24-built binary.
var PatchesFS embed.FS