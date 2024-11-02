package ast

import (
	"context"
	"fmt"
	"go/token"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/golang.org/x/tools/go/packages"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/util"
)

type LoadOptions struct {
	ForTest    bool
	Modes      []LoadMode
	BuildFlags []string // see FlagBuilder
}

func LoadPackages(ctx context.Context, dir string, args []string, opts *LoadOptions) ([]*packages.Package, *token.FileSet, error) {
	if dir == "" {
		return nil, nil, fmt.Errorf("requires dir")
	}
	absDir, err := util.ToAbsPath(dir)
	if err != nil {
		return nil, nil, err
	}

	pkgMode, err := LoadModes(opts.Modes).ToPackageMode()
	if err != nil {
		return nil, nil, err
	}
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Dir:        absDir,
		Mode:       pkgMode,
		Fset:       fset,
		Tests:      opts.ForTest,
		BuildFlags: opts.BuildFlags,
	}
	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		return nil, nil, err
	}
	return pkgs, fset, nil
}

func LoadSyntaxOnly(ctx context.Context, dir string, args []string, buildFlags []string) (absDir string, fset *token.FileSet, pkgs []*packages.Package, err error) {
	absDir, err = util.ToAbsPath(dir)
	if err != nil {
		return
	}
	pkgs, fset, err = LoadPackages(ctx, absDir, args, &LoadOptions{
		Modes:      []LoadMode{LoadMode_NeedSyntax, LoadMode_NeedName},
		BuildFlags: buildFlags,
	})
	return
}

type LoadModes []LoadMode

func (c LoadModes) ToPackageMode() (packages.LoadMode, error) {
	var m packages.LoadMode
	for _, loadMode := range c {
		pmode, ok := ModeMap[loadMode]
		if !ok {
			return 0, fmt.Errorf("load mode: %s not found", loadMode)
		}
		m |= pmode
	}
	return m, nil
}
