//go:build go1.24

package asset

import "embed"

// NOTE: "all:" prefix is required to include files/dirs whose names
// start with "_" or "." (e.g. __config__.json). Without "all:", go:embed
// silently excludes them.
//
//go:embed all:patches
var PatchesFS embed.FS