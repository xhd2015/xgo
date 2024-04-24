package goinfo

import (
	"os"
	"strings"
	"testing"
)

func TestParseVendorInfo(t *testing.T) {
	// NOTE: in vendor, replacing module have no
	// meta info
	info := ParseVendorContent(`# x.y/zz v1.0.2
## explicit; go 1.13
x.y/zz
# git.some/where/top-k v1.2.3-0.202222-cczz08
## explicit; go 1.19
git.some/where/top-k/boot
# git.some/x1/y1 v1.0.4 => git.some/x2/y2 v1.0.10
## explicit; go 1.14
git.some/x1/y1/mark
# git.some/x/y => git.some/x/y v1.3.0
# git.some/x1/y1 => git.some/x2/y2 v4.0.10
# git.some/x/sys => git.some/x/sys v0.0.0-20211216021012-1d35b9e2eb4e
`)
	expectVendors := []ModVersion{
		{Path: "x.y/zz", Version: "v1.0.2"},
		{Path: "git.some/where/top-k", Version: "v1.2.3-0.202222-cczz08"},
	}
	for i, expectVendor := range expectVendors {
		if info.VendorList[i] != expectVendor {
			t.Fatalf("expect %#v, actual: %#v", expectVendor, info.VendorList[i])
		}
	}
	t.Logf("%v", info)
}

func TestDebugParseCustomVendor(t *testing.T) {
	t.Skipf("debug only")

	debugPkg := "test"
	file := os.Getenv("TEST_DEBUG_FILE")

	vendorInfo, err := ParseVendor(file)
	if err != nil {
		t.Fatal(err)
	}

	for _, mod := range vendorInfo.VendorList {
		if strings.HasPrefix(debugPkg, mod.Path) {
			t.Logf("VendorList found mod: %s %s", mod.Path, mod.Version)
		}
	}
	for _, mod := range vendorInfo.VendorReplaced {
		if strings.HasPrefix(debugPkg, mod.Path) {
			t.Logf("VendorReplaced found mod: %s %s", mod.Path, mod.Version)
		}
	}
	for mod, meta := range vendorInfo.VendorMeta {
		if strings.HasPrefix(debugPkg, mod.Path) {
			t.Logf("VendorMeta found mod: %s %s, meta: %+v", mod.Path, mod.Version, meta)
		}
	}
	modVersion := vendorInfo.VendorVersion[debugPkg]
	t.Logf("VendorVersion: %s", modVersion)
}
