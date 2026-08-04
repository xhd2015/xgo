package overlay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComposeGoOverlays(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		caller    Replace
		generated Replace
		want      Replace
	}{
		{
			name: "no overlays",
			want: Replace{},
		},
		{
			name:   "caller mapping remains when not instrumented",
			caller: Replace{"A": "B"},
			want:   Replace{"A": "B"},
		},
		{
			name:      "caller replacement is instrumented",
			caller:    Replace{"A": "B"},
			generated: Replace{"B": "C"},
			want:      Replace{"A": "C", "B": "C"},
		},
		{
			name:      "generated replacement of caller source wins",
			caller:    Replace{"A": "B"},
			generated: Replace{"A": "C"},
			want:      Replace{"A": "C"},
		},
		{
			name:      "multi hop and unrelated generated mappings",
			caller:    Replace{"A": "B"},
			generated: Replace{"B": "C", "X": "Y", "Y": "Z"},
			want:      Replace{"A": "C", "B": "C", "X": "Z", "Y": "Z"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var caller *ParsedOverlay
			if tt.caller != nil {
				caller = ParseGoOverlay(&GoOverlay{Replace: tt.caller}, PathOptions{})
			}
			var generated *GoOverlay
			if tt.generated != nil {
				generated = &GoOverlay{Replace: tt.generated}
			}
			got, err := ComposeGoOverlays(caller, generated, PathOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Replace) != len(tt.want) {
				t.Fatalf("Replace length = %d, want %d: %#v", len(got.Replace), len(tt.want), got.Replace)
			}
			for source, want := range tt.want {
				if got.Replace[source] != want {
					t.Fatalf("Replace[%q] = %q, want %q", source, got.Replace[source], want)
				}
			}
		})
	}
}

func TestParseGoOverlayCanonical(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"source.go", "target.go"} {
		if err := os.WriteFile(filepath.Join(realDir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	input := &GoOverlay{Replace: Replace{
		AbsFile(filepath.Join(linkDir, "source.go")): AbsFile(filepath.Join(linkDir, "target.go")),
	}}

	parsed := ParseGoOverlay(input, PathOptions{})
	got := parsed.CanonicalReplace()
	wantSourcePath, err := filepath.EvalSymlinks(filepath.Join(realDir, "source.go"))
	if err != nil {
		t.Fatal(err)
	}
	wantTargetPath, err := filepath.EvalSymlinks(filepath.Join(realDir, "target.go"))
	if err != nil {
		t.Fatal(err)
	}
	wantSource := AbsFile(filepath.ToSlash(wantSourcePath))
	wantTarget := AbsFile(filepath.ToSlash(wantTargetPath))
	if got[wantSource] != wantTarget {
		t.Fatalf("canonical mapping = %#v, want %q -> %q", got, wantSource, wantTarget)
	}
	if len(parsed.Entries) != 1 {
		t.Fatalf("Entries length = %d, want 1", len(parsed.Entries))
	}
	e := parsed.Entries[0]
	if e.SourceResolved == e.SourceCanonical {
		// On some platforms the link path already equals the real path.
		return
	}
	// One pass records both spellings without a second overlay walk.
	if e.SourceCanonical != wantSource {
		t.Fatalf("SourceCanonical = %q, want %q", e.SourceCanonical, wantSource)
	}
	if e.TargetCanonical != wantTarget {
		t.Fatalf("TargetCanonical = %q, want %q", e.TargetCanonical, wantTarget)
	}
}

func TestParseGoOverlayPrefersCanonicalSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "source.go")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	canonical := CanonicalizeAbsFile(AbsFile(file), PathOptions{})
	alias := AbsFile(filepath.Join(filepath.Dir(file), "..", filepath.Base(filepath.Dir(file)), filepath.Base(file)))
	parsed := ParseGoOverlay(&GoOverlay{Replace: Replace{
		alias:     "caller.go",
		canonical: "generated.go",
	}}, PathOptions{})
	got := parsed.CanonicalReplace()
	if got[canonical] != "generated.go" {
		t.Fatalf("canonical source mapping = %q, want generated.go", got[canonical])
	}
}

func TestComposeGoOverlaysResolvesRelativePathsFromBaseDir(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	for _, name := range []string{
		"subject.go",
		"replacement/subject.go",
		".xgo/gen/subject.go",
	} {
		path := filepath.Join(baseDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	opts := PathOptions{BaseDir: baseDir}
	caller := ParseGoOverlay(&GoOverlay{Replace: Replace{"subject.go": "replacement/subject.go"}}, opts)
	generated := &GoOverlay{Replace: Replace{"./replacement/subject.go": ".xgo/gen/subject.go"}}

	got, err := ComposeGoOverlays(caller, generated, opts)
	if err != nil {
		t.Fatal(err)
	}
	wantSource := CanonicalizeAbsFile("subject.go", opts)
	wantTarget := CanonicalizeAbsFile(".xgo/gen/subject.go", opts)
	if got.Replace[wantSource] != wantTarget {
		t.Fatalf("composed mapping = %#v, want %q -> %q", got.Replace, wantSource, wantTarget)
	}

	if len(caller.Entries) != 1 {
		t.Fatalf("caller entries = %d, want 1", len(caller.Entries))
	}
	wantRawSource := AbsFile(filepath.ToSlash(filepath.Join(baseDir, "subject.go")))
	wantRawTarget := AbsFile(filepath.ToSlash(filepath.Join(baseDir, "replacement/subject.go")))
	if caller.Entries[0].SourceResolved != wantRawSource {
		t.Fatalf("SourceResolved = %q, want %q", caller.Entries[0].SourceResolved, wantRawSource)
	}
	if caller.Entries[0].TargetResolved != wantRawTarget {
		t.Fatalf("TargetResolved = %q, want %q", caller.Entries[0].TargetResolved, wantRawTarget)
	}
}

func TestProjectOntoCallerSources(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	for _, name := range []string{"subject.go", "replacement/subject.go", "gen/subject.go"} {
		path := filepath.Join(baseDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	opts := PathOptions{BaseDir: baseDir}
	caller := ParseGoOverlay(&GoOverlay{Replace: Replace{"subject.go": "replacement/subject.go"}}, opts)
	genTarget := CanonicalizeAbsFile("gen/subject.go", opts)
	composed := &GoOverlay{Replace: Replace{
		caller.Entries[0].SourceCanonical: genTarget,
	}}

	caller.ProjectOntoCallerSources(composed)
	if composed.Replace[caller.Entries[0].SourceResolved] != genTarget {
		t.Fatalf("projected mapping = %#v, want resolved source -> %q", composed.Replace, genTarget)
	}
}

func TestApplyFileRedirectsSingleIdentity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink: %v", err)
	}
	for _, name := range []string{"source.go", "target.go"} {
		if err := os.WriteFile(filepath.Join(realDir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	srcLink := AbsFile(filepath.Join(linkDir, "source.go"))
	tgtLink := AbsFile(filepath.Join(linkDir, "target.go"))
	srcReal := AbsFile(filepath.Join(realDir, "source.go"))
	parsed := ParseGoOverlay(&GoOverlay{Replace: Replace{srcLink: tgtLink}}, PathOptions{})
	if len(parsed.Entries) != 1 {
		t.Fatalf("entries = %d", len(parsed.Entries))
	}
	fs := MakeOverlay()
	parsed.ApplyFileRedirects(fs)
	// One FS entry; both spellings resolve via absFileKey.
	if len(fs) != 1 {
		t.Fatalf("overlay entries = %d, want 1", len(fs))
	}
	if fs.Get(srcLink) == nil {
		t.Fatalf("Get(link spelling) miss")
	}
	if fs.Get(srcReal) == nil {
		t.Fatalf("Get(real spelling) miss")
	}
}

func TestComposeGoOverlaysRejectsCycles(t *testing.T) {
	t.Parallel()

	_, err := ComposeGoOverlays(nil, &GoOverlay{Replace: Replace{"A": "B", "B": "A"}}, PathOptions{})
	if err == nil {
		t.Fatal("expected cycle error")
	}
	if !strings.Contains(err.Error(), "go overlay replacement cycle") {
		t.Fatalf("unexpected error: %v", err)
	}
}
