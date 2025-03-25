package instrument

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/instrument/inject"
	"github.com/xhd2015/xgo/support/instrument/overlay"
	"github.com/xhd2015/xgo/support/instrument/runtime"
)

// create an overlay: abs file -> content
type Overlay map[string]string

func LinkRuntime(projectRoot string, overlayFS overlay.Overlay) error {
	files, err := goinfo.ListFiles(projectRoot, []string{inject.XGO_RUNTIME_PKG})
	if err != nil {
		// TODO: handle the case where error indicates the package is not found
		return err
	}
	var linkFile string
	for _, file := range files {
		if strings.HasSuffix(file, inject.LINK_FILE) {
			linkFile = file
			break
		}
	}
	if linkFile == "" {
		return fmt.Errorf("link file not found")
	}
	overlayFS.OverrideContent(overlay.AbsFile(linkFile), inject.GetLinkRuntimeCode())
	return nil
}

func InstrumentUserCode(projectRoot string, overlayFS overlay.Overlay, buildArgs []string) error {
	projectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return err
	}

	files, err := goinfo.ListRelativeFiles(projectRoot, buildArgs)
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		absFile := overlay.AbsFile(filepath.Join(projectRoot, file))
		content, ok, err := inject.InjectRuntimeTrap(absFile, overlayFS)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		overlayFS.OverrideContent(absFile, string(content))
	}

	return nil
}

func InstrumentRuntime(goroot string, opts runtime.InstrumentRuntimeOptions) error {
	return runtime.InstrumentRuntime(goroot, opts)
}
