package overlay

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/fileutil"
)

type AbsFile string

// serves as an in-memory FS
type Overlay map[AbsFile]*FileOverlay

type FileOverlay struct {
	AbsFile AbsFile
	Content string

	hasOverriddenContent bool
}

func MakeOverlay() Overlay {
	return make(Overlay)
}

// absFileKey is the single identity used inside Overlay.
// Paths are kept in native filepath form (no ToSlash) so lookups match
// package-load AbsPaths on Windows, and EvalSymlinks so /var and
// /private/var on macOS share one entry. Callers may pass either spelling;
// Get/Override always store and look up under this key.
func absFileKey(file AbsFile) AbsFile {
	path := filepath.FromSlash(string(file))
	path = filepath.Clean(path)
	if resolved, err := evalSymlinks(path); err == nil {
		path = resolved
	}
	return AbsFile(path)
}

func (o Overlay) OverrideFile(absFile AbsFile, targetFile AbsFile) {
	o[absFileKey(absFile)] = &FileOverlay{
		AbsFile: absFileKey(targetFile),
	}
}

func (o Overlay) OverrideContent(absFile AbsFile, content string) {
	// Replaces any prior file redirect for the same identity, so a later
	// instrumented content mapping cannot leave an uninstrumented alias behind.
	o[absFileKey(absFile)] = &FileOverlay{
		Content:              content,
		hasOverriddenContent: true,
	}
}

func (o Overlay) Get(absFile AbsFile) *FileOverlay {
	return o[absFileKey(absFile)]
}

func (o Overlay) Size(absFile AbsFile) (size int64, err error) {
	overlayFile := o.Get(absFile)

	readOSFile := absFile
	if overlayFile != nil {
		if overlayFile.hasOverriddenContent {
			return int64(len(overlayFile.Content)), nil
		}
		if overlayFile.AbsFile != "" {
			readOSFile = overlayFile.AbsFile
		}
	}
	info, err := os.Stat(string(readOSFile))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (o Overlay) Read(absFile AbsFile) (hitContent bool, content string, err error) {
	overlayFile := o.Get(absFile)

	readOSFile := absFile
	if overlayFile != nil {
		if overlayFile.hasOverriddenContent {
			return true, overlayFile.Content, nil
		}
		if overlayFile.AbsFile != "" {
			readOSFile = overlayFile.AbsFile
		}
	}
	data, err := os.ReadFile(string(readOSFile))
	if err != nil {
		return false, "", err
	}
	return false, string(data), nil
}

func (o Overlay) ReadBytes(absFile AbsFile) (hitContent bool, content []byte, err error) {
	overlayFile := o.Get(absFile)

	readOSFile := absFile
	if overlayFile != nil {
		if overlayFile.hasOverriddenContent {
			return true, []byte(overlayFile.Content), nil
		}
		if overlayFile.AbsFile != "" {
			readOSFile = overlayFile.AbsFile
		}
	}
	data, err := os.ReadFile(string(readOSFile))
	if err != nil {
		return false, nil, err
	}
	return false, data, nil
}

type Options struct {
	NoLineDirective bool
	PathMappings    []PathMapping
}

// PROJECT
// GOROOT
// GOPATH
type PathMapping struct {
	From string
	To   string
}

func getPathMapping(path string, mappings []PathMapping) string {
	for _, mapping := range mappings {
		// normalize to forward slashes for cross-platform matching
		normPath := filepath.ToSlash(path)
		normFrom := filepath.ToSlash(mapping.From)
		if !strings.HasPrefix(normPath, normFrom) {
			continue
		}
		if len(normPath) == len(normFrom) {
			return mapping.To
		}
		if normPath[len(normFrom)] != '/' {
			continue
		}
		return mapping.To + normPath[len(normFrom):]
	}
	return path
}

// MakeGoOverlay writes the overridden contents under overlayDir, each at
// its rebased path below a directory named after its content: xgo runs in
// the same project, e.g. for different packages, generate different
// contents for a file, and must not overwrite the one another run's go
// command is reading. A file is written only when no run wrote that
// content before, through a rename so that it is never seen partially
// written, and touched otherwise so that CleanGoOverlay keeps it.
func (o Overlay) MakeGoOverlay(overlayDir string, opts Options) (*GoOverlay, error) {
	noLineDirective := opts.NoLineDirective
	pathMappings := opts.PathMappings

	absOverlayDir, err := filepath.Abs(overlayDir)
	if err != nil {
		return nil, err
	}
	replace := make(Replace, len(o))
	type writeJob struct {
		rebased string
		absFile AbsFile
		content string
		file    string // set by write
	}
	var jobs []*writeJob
	for absFile, fileOverlay := range o {
		if fileOverlay.hasOverriddenContent {
			absFilePath := getPathMapping(string(absFile), pathMappings)
			jobs = append(jobs, &writeJob{
				rebased: fileutil.RebaseAbsPath(absOverlayDir, absFilePath),
				absFile: absFile,
				content: fileOverlay.Content,
			})
			continue
		}
		if fileOverlay.AbsFile == "" {
			continue
		}
		replace[AbsFile(filepath.ToSlash(string(absFile)))] = AbsFile(filepath.ToSlash(string(fileOverlay.AbsFile)))
	}
	// a large module has thousands of instrumented files: write them in
	// parallel
	write := func(job *writeJob) error {
		var prefix string
		if !noLineDirective {
			prefix = patch.FmtLineDirective(string(job.absFile), 1) + "\n"
		}
		data := make([]byte, 0, len(prefix)+len(job.content))
		data = append(data, prefix...)
		data = append(data, job.content...)
		var err error
		job.file, err = WriteContentFile(filepath.Dir(job.rebased), filepath.Base(job.rebased), data)
		return err
	}
	workers := runtime.GOMAXPROCS(0)
	if workers > len(jobs) {
		workers = len(jobs)
	}
	var wg sync.WaitGroup
	errs := make([]error, workers)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := w; i < len(jobs); i += workers {
				if err := write(jobs[i]); err != nil {
					errs[w] = err
					return
				}
			}
		}(w)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	for _, job := range jobs {
		replace[AbsFile(filepath.ToSlash(string(job.absFile)))] = AbsFile(filepath.ToSlash(job.file))
	}
	return &GoOverlay{Replace: replace}, nil
}
