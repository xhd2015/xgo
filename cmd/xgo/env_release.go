//go:build ignore

package main

import "embed"

const isDevelopment = false

//go:embed patch_compiler
var patchEmbed embed.FS
