package cov_control

import (
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/coverage"
	"github.com/xhd2015/xgo/support/fileutil"
)

type Controller interface {
	GetCoverage() (*coverage.Detail, error)
	GetCoverageProfilePath() (string, error)
	Reset() error
	Refresh() error
}

func New(projectDir string) (Controller, error) {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}
	covDir := mapCovDir(getTmpDir(), absDir)
	return &controller{
		covDir:          covDir,
		projectDir:      projectDir,
		absProjectDir:   absDir,
		coverageProfile: filepath.Join(covDir, "coverage.out"),
	}, nil
}

type controller struct {
	covDir          string
	projectDir      string
	absProjectDir   string
	coverageProfile string
}

func (c *controller) GetCoverage() (*coverage.Detail, error) {
	panic("unimplemented")
}

func (c *controller) GetCoverageProfilePath() (string, error) {
	return c.coverageProfile, nil
}

func (c *controller) Refresh() error {
	panic("unimplemented")
}

func (c *controller) Reset() error {
	err := os.RemoveAll(c.coverageProfile)
	if err != nil {
		return err
	}
	return c.Refresh()
}

func mapCovDir(tmpDir string, absDir string) string {
	return filepath.Join(tmpDir, "xgo", "coverage", fileutil.CleanSpecial(absDir))
}

func getTmpDir() string {
	// try /tmp first for shorter path
	tmpDir := "/tmp"
	stat, statErr := os.Stat(tmpDir)
	if statErr == nil && stat.IsDir() {
		return tmpDir
	}
	return os.TempDir()
}
