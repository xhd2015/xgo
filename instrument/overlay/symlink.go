package overlay

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// resolvedDirs caches filepath.EvalSymlinks of directories. Only results
// are cached, not errors: xgo creates directories while it runs.
var resolvedDirs sync.Map // string -> string

// evalSymlinks is filepath.EvalSymlinks with the directory of path resolved
// once: every overlay lookup canonicalizes its file, and resolving the
// whole path took one lstat per component, for each lookup of each file.
func evalSymlinks(path string) (string, error) {
	path = filepath.Clean(path)
	// on windows EvalSymlinks also normalizes the case of each component
	if runtime.GOOS == "windows" || !filepath.IsAbs(path) {
		return filepath.EvalSymlinks(path)
	}
	dir := filepath.Dir(path)
	if dir == path {
		return filepath.EvalSymlinks(path)
	}
	resolvedDir, err := resolveDir(dir)
	if err != nil {
		return "", err
	}
	file := filepath.Join(resolvedDir, filepath.Base(path))
	info, err := os.Lstat(file)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return filepath.EvalSymlinks(file)
	}
	return file, nil
}

// resolveDir is filepath.EvalSymlinks of the clean absolute directory dir,
// resolved from its cached parent: the generated files live in one
// directory each (named after their content), and resolving each of them
// from the root took one lstat per component.
func resolveDir(dir string) (string, error) {
	if v, ok := resolvedDirs.Load(dir); ok {
		return v.(string), nil
	}
	parent := filepath.Dir(dir)
	if parent == dir {
		return filepath.EvalSymlinks(dir)
	}
	resolvedParent, err := resolveDir(parent)
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(resolvedParent, filepath.Base(dir))
	info, err := os.Lstat(resolved)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		resolved, err = filepath.EvalSymlinks(resolved)
		if err != nil {
			return "", err
		}
	}
	resolvedDirs.Store(dir, resolved)
	return resolved, nil
}
