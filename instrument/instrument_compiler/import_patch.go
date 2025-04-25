package instrument_compiler

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/cmd/xgo/asset"
	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/embed"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/osinfo"
)

var XgoRewriteInternal = patch.FilePath{"src", "cmd", "compile", "internal", "xgo_rewrite_internal"}
var XgoRewriteInternalPatch = append(XgoRewriteInternal, "patch")

func ImportCompileInternalPatch(goroot string, xgoSrc string, forceReset bool, syncWithLink bool) error {
	dstDir := JoinInternalPatch(goroot)
	if config.IS_DEV {
		symLink := syncWithLink
		if osinfo.FORCE_COPY_UNSYM {
			// Windows: A required privilege is not held by the client.
			symLink = false
		}
		copier := filecopy.NewOptions().Ignore("legacy")
		if symLink {
			copier.UseLink()
		}
		// copy compiler internal dependencies
		err := copier.CopyReplaceDir(filepath.Join(xgoSrc, "patch"), dstDir)
		if err != nil {
			return err
		}

		// remove patch/go.mod
		err = os.RemoveAll(filepath.Join(dstDir, "go.mod"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		return nil
	}

	if forceReset {
		// -a causes repatch
		err := os.RemoveAll(dstDir)
		if err != nil {
			return err
		}
	} else {
		// check if already copied
		_, statErr := os.Stat(dstDir)
		if statErr == nil {
			// skip copy if already exists
			return nil
		}
	}

	// read from embed
	err := embed.CopyDir(asset.CompilerPatchGenFS, asset.CompilerPatchGen, dstDir, embed.CopyOptions{
		IgnorePaths: []string{"go.mod", "legacy"},
	})
	if err != nil {
		return err
	}

	return nil
}

func JoinInternalPatch(goroot string, subDirs ...string) string {
	dir := filepath.Join(goroot, filepath.Join(XgoRewriteInternalPatch...))
	if len(subDirs) > 0 {
		dir = filepath.Join(dir, filepath.Join(subDirs...))
	}
	return dir
}
