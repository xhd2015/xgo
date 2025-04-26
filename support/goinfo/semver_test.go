package goinfo

import "testing"

func TestCompareSemVer(t *testing.T) {
	testCases := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		// Basic version comparisons
		{"major version", "v1.0.0", "v2.0.0", -1},
		{"minor version", "v1.1.0", "v1.2.0", -1},
		{"patch version", "v1.0.1", "v1.0.2", -1},
		{"equal versions", "v1.0.0", "v1.0.0", 0},
		{"reverse order", "v2.0.0", "v1.0.0", 1},

		// Pre-release versions
		{"pre-release vs release", "v1.0.0-alpha", "v1.0.0", -1},
		{"pre-release comparison", "v1.0.0-alpha", "v1.0.0-beta", -1},
		{"pre-release numeric", "v1.0.0-alpha.1", "v1.0.0-alpha.2", -1},
		{"pre-release longer", "v1.0.0-alpha", "v1.0.0-alpha.1", -1},

		// Invalid versions
		{"invalid vs valid", "invalid", "v1.0.0", -1},
		{"valid vs invalid", "v1.0.0", "invalid", 1},
		{"both invalid", "invalid1", "invalid2", 0},
		{"empty vs valid", "", "v1.0.0", -1},
		{"valid vs empty", "v1.0.0", "", 1},
		{"both empty", "", "", 0},

		// Edge cases
		{"incomplete version", "v1.0", "v1.0.0", 0},     // v1.0 is treated as v1.0.0
		{"major only", "v1", "v1.0.0", 0},               // v1 is treated as v1.0.0
		{"different incomplete", "v1.0", "v2.0", -1},    // v1.0 vs v2.0
		{"no v prefix", "1.0.0", "v1.0.0", -1},          // 1.0.0 is invalid (no v prefix)
		{"build metadata", "v1.0.0+build", "v1.0.0", 0}, // build metadata doesn't affect comparison
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := CompareSemVer(tc.v1, tc.v2)
			if got != tc.want {
				t.Errorf("CompareSemVer(%q, %q) = %d, want %d", tc.v1, tc.v2, got, tc.want)
			}
		})
	}
}
