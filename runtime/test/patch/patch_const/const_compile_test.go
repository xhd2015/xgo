package patch_const

import "testing"

const XGO_VERSION = ""

// see https://github.com/xhd2015/xgo/issues/78
func TestConstCompile(t *testing.T) {
	// all operators
	if XGO_VERSION != "" {
		t.Fatalf("fail")
	}
	if XGO_VERSION < "" {
		t.Fatalf("fail")
	}

	if XGO_VERSION > "" {
		t.Fatalf("fail")
	}
	if XGO_VERSION <= "" {
	} else {
		t.Fatalf("fail")
	}

	if XGO_VERSION >= "" {
	} else {
		t.Fatalf("fail")
	}

	if XGO_VERSION == "" {
	} else {
		t.Fatalf("fail")
	}
	if "" != XGO_VERSION {
		t.Fatalf("fail")
	}
}
