// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goinfo

import (
	"os"
	"path/filepath"
	"strings"
)

// source code from https://cs.opensource.google/go/go/+/master:src/cmd/go/internal/modload/vendor.go

type vendorMetadata struct {
	Explicit    bool
	Replacement ModVersion
	GoVersion   string
}
type ModVersion struct {
	Path    string
	Version string
}

// ParseVendor reads the list of vendored modules from vendor/modules.txt.
func ParseVendor(dirOrFile string) (*VendorInfo, error) {
	// vendorPkgModule = make(map[string]ModVersion)
	// vendorVersion = make(map[string]string)
	// vendorMeta = make(map[ModVersion]vendorMetadata)
	vendorFile := dirOrFile
	if !strings.HasSuffix(dirOrFile, "modules.txt") {
		vendorFile = filepath.Join(dirOrFile, "modules.txt")
	}
	data, err := os.ReadFile(vendorFile)
	if err != nil {
		return nil, err
	}

	return ParseVendorContent(string(data)), nil
}

func ParseVendorContent(content string) *VendorInfo {
	info := &VendorInfo{
		VendorVersion:   make(map[string]string),
		VendorPkgModule: make(map[string]ModVersion),
		VendorMeta:      make(map[ModVersion]vendorMetadata),
	}
	info.parseContent(content)
	return info
}

type VendorInfo struct {
	VendorList      []ModVersion          // modules that contribute packages to the build, in order of appearance
	VendorReplaced  []ModVersion          // all replaced modules; may or may not also contribute packages
	VendorVersion   map[string]string     // module path → selected version (if known)
	VendorPkgModule map[string]ModVersion // package → containing module
	VendorMeta      map[ModVersion]vendorMetadata

	// version on otp
	mod ModVersion
}

func (c *VendorInfo) parseContent(content string) {
	for _, line := range strings.Split(content, "\n") {
		c.parseLine(line)
	}
}

func (c *VendorInfo) parseLine(line string) {
	vendorMeta := c.VendorMeta

	if strings.HasPrefix(line, "# ") {
		f := strings.Fields(line)

		if len(f) < 3 {
			return
		}
		if IsValidSemVer(f[2]) {
			// A module, but we don't yet know whether it is in the build list or
			// only included to indicate a replacement.
			c.mod = ModVersion{Path: f[1], Version: f[2]}
			f = f[3:]
		} else if f[2] == "=>" {
			// A wildcard replacement found in the main module's go.mod file.
			c.mod = ModVersion{Path: f[1]}
			f = f[2:]
		} else {
			// Not a version or a wildcard replacement.
			// We don't know how to interpret this module line, so ignore it.
			c.mod = ModVersion{}
			return
		}

		if len(f) >= 2 && f[0] == "=>" {
			meta := vendorMeta[c.mod]
			if len(f) == 2 {
				// File replacement.
				meta.Replacement = ModVersion{Path: f[1]}
				c.VendorReplaced = append(c.VendorReplaced, c.mod)
			} else if len(f) == 3 && IsValidSemVer(f[2]) {
				// Path and version replacement.
				meta.Replacement = ModVersion{Path: f[1], Version: f[2]}
				c.VendorReplaced = append(c.VendorReplaced, c.mod)
			} else {
				// We don't understand this replacement. Ignore it.
			}
			vendorMeta[c.mod] = meta
		}
		return
	}

	// Not a module line. Must be a package within a module or a metadata
	// directive, either of which requires a preceding module line.
	if c.mod.Path == "" {
		return
	}

	if annotations, ok := strings.CutPrefix(line, "## "); ok {
		// Metadata. Take the union of annotations across multiple lines, if present.
		meta := vendorMeta[c.mod]
		for _, entry := range strings.Split(annotations, ";") {
			entry = strings.TrimSpace(entry)
			if entry == "explicit" {
				meta.Explicit = true
			}
			if goVersion, ok := strings.CutPrefix(entry, "go "); ok {
				meta.GoVersion = goVersion
			}
			// All other tokens are reserved for future use.
		}
		vendorMeta[c.mod] = meta
		return
	}

	// assume f[0] is valid path
	if f := strings.Fields(line); len(f) == 1 {
		// A package within the current module.
		c.VendorPkgModule[f[0]] = c.mod

		// Since this module provides a package for the build, we know that it
		// is in the build list and is the selected version of its path.
		// If this information is new, record it.
		if v, ok := c.VendorVersion[c.mod.Path]; !ok || ModCompare(c.mod.Path, v, c.mod.Version) < 0 {
			c.VendorList = append(c.VendorList, c.mod)
			c.VendorVersion[c.mod.Path] = c.mod.Version
		}
	}
}
