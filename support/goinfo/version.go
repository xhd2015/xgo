// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gover implements support for Go toolchain versions like 1.21.0 and 1.21rc1.
// (For historical reasons, Go does not use semver for its toolchains.)
// This package provides the same basic analysis that golang.org/x/mod/semver does for semver.
//
// The go/version package should be imported instead of this one when possible.
// Note that this package works on "1.21" while go/version works on "go1.21".

package goinfo

// see: https://cs.opensource.google/go/go/+/master:src/internal/gover/gover.go
type Version struct {
	Major string // decimal
	Minor string // decimal or ""
	Patch string // decimal or ""
	Kind  string // "", "alpha", "beta", "rc"
	Pre   string // decimal or ""
}

func ParseVersion(x string) Version {
	var v Version

	// Parse major version.
	var ok bool
	v.Major, x, ok = cutInt(x)
	if !ok {
		return Version{}
	}
	if x == "" {
		// Interpret "1" as "1.0.0".
		v.Minor = "0"
		v.Patch = "0"
		return v
	}

	// Parse . before minor version.
	if x[0] != '.' {
		return Version{}
	}

	// Parse minor version.
	v.Minor, x, ok = cutInt(x[1:])
	if !ok {
		return Version{}
	}
	if x == "" {
		// Patch missing is same as "0" for older versions.
		// Starting in Go 1.21, patch missing is different from explicit .0.
		if CmpInt(v.Minor, "21") < 0 {
			v.Patch = "0"
		}
		return v
	}

	// Parse patch if present.
	if x[0] == '.' {
		v.Patch, x, ok = cutInt(x[1:])
		if !ok || x != "" {
			// Note that we are disallowing prereleases (alpha, beta, rc) for patch releases here (x != "").
			// Allowing them would be a bit confusing because we already have:
			//	1.21 < 1.21rc1
			// But a prerelease of a patch would have the opposite effect:
			//	1.21.3rc1 < 1.21.3
			// We've never needed them before, so let's not start now.
			return Version{}
		}
		return v
	}

	// Parse prerelease.
	i := 0
	for i < len(x) && (x[i] < '0' || '9' < x[i]) {
		if x[i] < 'a' || 'z' < x[i] {
			return Version{}
		}
		i++
	}
	if i == 0 {
		return Version{}
	}
	v.Kind, x = x[:i], x[i:]
	if x == "" {
		return v
	}
	v.Pre, x, ok = cutInt(x)
	if !ok || x != "" {
		return Version{}
	}

	return v
}
func CompareVersion(x, y string) int {
	vx := ParseVersion(x)
	vy := ParseVersion(y)

	if c := CmpInt(vx.Major, vy.Major); c != 0 {
		return c
	}
	if c := CmpInt(vx.Minor, vy.Minor); c != 0 {
		return c
	}
	if c := CmpInt(vx.Patch, vy.Patch); c != 0 {
		return c
	}
	if c := compareString(vx.Kind, vy.Kind); c != 0 { // "" < alpha < beta < rc
		return c
	}
	if c := CmpInt(vx.Pre, vy.Pre); c != 0 {
		return c
	}
	return 0
}
func IsValidVersion(x string) bool {
	return ParseVersion(x) != Version{}
}

func compareString(x, y string) int {
	if x < y {
		return -1
	}
	if x > y {
		return +1
	}
	return 0
}

func CmpInt(x, y string) int {
	if x == y {
		return 0
	}
	if len(x) < len(y) {
		return -1
	}
	if len(x) > len(y) {
		return +1
	}
	if x < y {
		return -1
	} else {
		return +1
	}
}

func cutInt(x string) (n, rest string, ok bool) {
	i := 0
	for i < len(x) && '0' <= x[i] && x[i] <= '9' {
		i++
	}
	if i == 0 || x[0] == '0' && i != 1 { // no digits or unnecessary leading zero
		return "", "", false
	}
	return x[:i], x[i:], true
}
