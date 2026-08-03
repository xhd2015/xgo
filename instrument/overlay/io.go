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

// CanonicalizeGoOverlay returns an in-memory copy with paths normalized to the
// filesystem identity used by the Go loader. In particular, macOS callers may
// supply /var/... while package loading reports /private/var/... for the same
// file. Keeping one identity lets xgo replace and later compose that file.
// The caller's JSON file is never modified.
func ResolveGoOverlayPaths(input *GoOverlay, opts PathOptions) *GoOverlay {
	if input == nil {
		return nil
	}
	result := &GoOverlay{Replace: make(Replace, len(input.Replace))}
	for source, target := range input.Replace {
		result.Replace[ResolveAbsFile(source, opts)] = ResolveAbsFile(target, opts)
	}
	return result
}

func ResolveAbsFile(file AbsFile, opts PathOptions) AbsFile {
	path := filepath.FromSlash(string(file))
	if !filepath.IsAbs(path) && opts.BaseDir != "" {
		path = filepath.Join(opts.BaseDir, path)
	}
	return AbsFile(filepath.ToSlash(filepath.Clean(path)))
}

func CanonicalizeGoOverlay(input *GoOverlay, opts PathOptions) *GoOverlay {
	if input == nil {
		return nil
	}
	result := &GoOverlay{Replace: make(Replace, len(input.Replace))}
	for source, target := range input.Replace {
		canonicalSource := CanonicalizeAbsFile(source, opts)
		// Prefer a mapping already expressed using the canonical source path.
		// This lets an instrumented mapping supersede an alias retained from a
		// caller overlay (for example /private/var over /var on macOS).
		if _, exists := result.Replace[canonicalSource]; !exists || source == canonicalSource {
			result.Replace[canonicalSource] = CanonicalizeAbsFile(target, opts)
		}
	}
	return result
}

func CanonicalizeAbsFile(file AbsFile, opts PathOptions) AbsFile {
	path := filepath.FromSlash(string(ResolveAbsFile(file, opts)))
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
func ComposeGoOverlays(caller, generated *GoOverlay, opts PathOptions) (*GoOverlay, error) {
	caller = CanonicalizeGoOverlay(caller, opts)
	generated = CanonicalizeGoOverlay(generated, opts)
	combined := make(Replace)
	if caller != nil {
		for source, target := range caller.Replace {
			combined[source] = target
		}
	}
	if generated != nil {
		for source, target := range generated.Replace {
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
