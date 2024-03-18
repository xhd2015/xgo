package main

import "embed"

const isDevelopment = true

// NOTE: patchEmbed will be rewritten when building release, see script/build-release/fixup.go for details
var patchEmbed embed.FS
