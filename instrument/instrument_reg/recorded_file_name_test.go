package instrument_reg

import "testing"

func TestRecordedFileName(t *testing.T) {
	abs := "/home/ci/slot-3/_work/repo/repo/pkg/shareid/shareid.go"
	if got := recordedFileName("github.com/acme/repo/pkg/shareid", abs, false); got != abs {
		t.Fatalf("without trimpath: got %q, want absolute path", got)
	}
	want := "github.com/acme/repo/pkg/shareid/shareid.go"
	if got := recordedFileName("github.com/acme/repo/pkg/shareid", abs, true); got != want {
		t.Fatalf("with trimpath: got %q, want %q", got, want)
	}
	// Two checkouts of the same tree must record the same name under -trimpath.
	other := "/opt/slot-4/_work/repo/repo/pkg/shareid/shareid.go"
	if recordedFileName("github.com/acme/repo/pkg/shareid", abs, true) != recordedFileName("github.com/acme/repo/pkg/shareid", other, true) {
		t.Fatal("trimpath names differ between checkouts")
	}
}
