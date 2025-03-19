package instrument

import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/instrument/inject"
	"github.com/xhd2015/xgo/support/instrument/runtime"
)

// create an overlay: file -> content
type Overlay map[string]string

func InstrumentUserCode(projectRoot string, buildArgs []string) (Overlay, error) {
	projectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, err
	}
	overlay := make(Overlay)

	files, err := goinfo.ListRelativeFiles(projectRoot, buildArgs)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		content, ok, err := inject.InjectRuntimeTrap(file)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		overlay[filepath.Join(projectRoot, file)] = string(content)
	}

	return overlay, nil
}

func InstrumentRuntime(goroot string) error {
	return runtime.InstrumentRuntime(goroot)
}
