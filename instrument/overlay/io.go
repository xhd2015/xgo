package overlay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Replace map[AbsFile]AbsFile

type GoOverlay struct {
	Replace Replace
}

// PathOptions defines how Go overlay paths are interpreted. BaseDir is the
// target project's directory, not the xgo process working directory.
type PathOptions struct {
	BaseDir string
}

// OverlayEntry is one caller Replace pair after a single normalize pass.
// Resolved paths are abs+clean (no EvalSymlinks). Canonical paths add
// EvalSymlinks so they match the filesystem identity used by package load.
type OverlayEntry struct {
	// SourceInput is the Replace key as provided in the overlay JSON.
	SourceInput AbsFile

	SourceResolved  AbsFile
	SourceCanonical AbsFile
	TargetResolved  AbsFile
	TargetCanonical AbsFile
}

// ParsedOverlay is the in-memory form of a Go -overlay after one parse pass.
// Callers apply dual path spellings via methods instead of keeping two maps.
type ParsedOverlay struct {
	Entries []OverlayEntry
}

func (c *GoOverlay) Write(file string) error {
	overlayData, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(file, overlayData, 0755)
}

func ReadGoOverlay(file string) (*GoOverlay, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var overlay GoOverlay
	err = json.Unmarshal(content, &overlay)
	if err != nil {
		return nil, err
	}
	return &overlay, nil
}

// ParseGoOverlay walks input.Replace once and records both resolved and
// canonical spellings for each mapping. The caller's JSON is never modified.
func ParseGoOverlay(input *GoOverlay, opts PathOptions) *ParsedOverlay {
	if input == nil {
		return nil
	}
	out := &ParsedOverlay{Entries: make([]OverlayEntry, 0, len(input.Replace))}
	for source, target := range input.Replace {
		srcRes := ResolveAbsFile(source, opts)
		tgtRes := ResolveAbsFile(target, opts)
		out.Entries = append(out.Entries, OverlayEntry{
			SourceInput:     source,
			SourceResolved:  srcRes,
			SourceCanonical: evalSymlinksAbs(srcRes),
			TargetResolved:  tgtRes,
			TargetCanonical: evalSymlinksAbs(tgtRes),
		})
	}
	return out
}

// ApplyFileRedirects registers file redirects into overlayFS.
// When resolved and canonical source spellings differ (e.g. /var vs
// /private/var on macOS), both keys are registered so either lookup hits.
func (p *ParsedOverlay) ApplyFileRedirects(fs Overlay) {
	if p == nil {
		return
	}
	for _, e := range p.Entries {
		fs.OverrideFile(e.SourceResolved, e.TargetResolved)
		if e.SourceCanonical != e.SourceResolved {
			fs.OverrideFile(e.SourceCanonical, e.TargetCanonical)
		}
	}
}

// CanonicalReplace returns the map used for compose and loader identity.
// When multiple entries collapse to the same canonical source, prefer a
// mapping whose input key is already the canonical spelling (so an explicit
// /private/var key supersedes a /var alias, matching prior behavior).
func (p *ParsedOverlay) CanonicalReplace() Replace {
	if p == nil {
		return nil
	}
	result := make(Replace, len(p.Entries))
	for _, e := range p.Entries {
		prefer := e.SourceInput == e.SourceCanonical
		if _, exists := result[e.SourceCanonical]; !exists || prefer {
			result[e.SourceCanonical] = e.TargetCanonical
		}
	}
	return result
}

// ProjectOntoCallerSources copies composed targets onto the resolved source
// spellings that Go's -overlay may see literally.
func (p *ParsedOverlay) ProjectOntoCallerSources(composed *GoOverlay) {
	if p == nil || composed == nil {
		return
	}
	if composed.Replace == nil {
		composed.Replace = make(Replace)
	}
	for _, e := range p.Entries {
		if target, ok := composed.Replace[e.SourceCanonical]; ok {
			composed.Replace[e.SourceResolved] = target
		}
	}
}

func ResolveAbsFile(file AbsFile, opts PathOptions) AbsFile {
	path := filepath.FromSlash(string(file))
	if !filepath.IsAbs(path) && opts.BaseDir != "" {
		path = filepath.Join(opts.BaseDir, path)
	}
	return AbsFile(filepath.ToSlash(filepath.Clean(path)))
}

// CanonicalizeAbsFile resolves and symlink-canonicalizes a single path.
// Prefer ParseGoOverlay when normalizing a full overlay so each path is
// resolved only once.
func CanonicalizeAbsFile(file AbsFile, opts PathOptions) AbsFile {
	return evalSymlinksAbs(ResolveAbsFile(file, opts))
}

func evalSymlinksAbs(file AbsFile) AbsFile {
	path := filepath.FromSlash(string(file))
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return AbsFile(filepath.ToSlash(filepath.Clean(path)))
}

// ComposeGoOverlays returns the single-level Go overlay obtained by applying
// generated after caller. xgo may instrument a caller replacement, so a
// caller mapping A -> B and generated mapping B -> C must be emitted as
// A -> C: the Go tool does not follow overlay mappings transitively.
//
// The returned overlay also retains generated mappings that have no caller
// origin. Cyclic mappings are rejected because they cannot be represented by
// a usable Go overlay.
//
// caller may be nil. generated is parsed in one pass when non-nil.
func ComposeGoOverlays(caller *ParsedOverlay, generated *GoOverlay, opts PathOptions) (*GoOverlay, error) {
	combined := make(Replace)
	if caller != nil {
		for source, target := range caller.CanonicalReplace() {
			combined[source] = target
		}
	}
	if generated != nil {
		// Generated paths (from MakeGoOverlay) are already abs; parse once so
		// relative keys and symlink aliases still share loader identity.
		for source, target := range ParseGoOverlay(generated, opts).CanonicalReplace() {
			// A generated mapping is xgo's instrumented version of the
			// effective source and must take precedence over the caller input.
			combined[source] = target
		}
	}

	resolved := make(Replace, len(combined))
	visiting := make(map[AbsFile]bool, len(combined))
	var resolve func(AbsFile, []AbsFile) (AbsFile, error)
	resolve = func(source AbsFile, path []AbsFile) (AbsFile, error) {
		if target, ok := resolved[source]; ok {
			return target, nil
		}
		if visiting[source] {
			cycle := append(path, source)
			parts := make([]string, len(cycle))
			for i, file := range cycle {
				parts[i] = string(file)
			}
			return "", fmt.Errorf("go overlay replacement cycle: %s", strings.Join(parts, " -> "))
		}
		target, ok := combined[source]
		if !ok {
			return source, nil
		}
		visiting[source] = true
		terminal, err := resolve(target, append(path, source))
		delete(visiting, source)
		if err != nil {
			return "", err
		}
		resolved[source] = terminal
		return terminal, nil
	}

	keys := make([]string, 0, len(combined))
	for source := range combined {
		keys = append(keys, string(source))
	}
	sort.Strings(keys)
	for _, source := range keys {
		if _, err := resolve(AbsFile(source), nil); err != nil {
			return nil, err
		}
	}
	return &GoOverlay{Replace: resolved}, nil
}
