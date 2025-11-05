package cov_control

// this controller is not used in practice, we may use it to
// optimize user experience in the future.
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/git"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/load/loadcov"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/coverage"
	"github.com/xhd2015/xgo/cmd/xgo/test-explorer/icov"
	"github.com/xhd2015/xgo/support/fileutil"
)

func New(projectDir string, args []string, diffBase string, profile string) (icov.Controller, error) {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}
	covDir := mapCovDir(getTmpDir(), absDir)
	if profile == "" {
		profile = filepath.Join(covDir, "coverage.out")
	}
	err = os.MkdirAll(covDir, 0755)
	if err != nil {
		return nil, err
	}
	return &controller{
		covDir:          covDir,
		args:            args,
		diffBase:        diffBase,
		projectDir:      projectDir,
		absProjectDir:   absDir,
		coverageProfile: profile,
	}, nil
}

type controller struct {
	covDir          string
	args            []string
	diffBase        string
	projectDir      string
	absProjectDir   string
	coverageProfile string
}

var _ icov.Controller = (*controller)(nil)

func (c *controller) GetCoverage() (*coverage.Detail, error) {
	resFile := c.cacheResultPath()
	data, err := os.ReadFile(resFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var detail coverage.Detail
	jsonErr := json.Unmarshal(data, &detail)
	if jsonErr != nil {
		return nil, nil
	}
	return &detail, nil
}

func (c *controller) GetCoverageProfilePath() (string, error) {
	return c.coverageProfile, nil
}

func (c *controller) Refresh() error {
	ok, err := fileExists(c.coverageProfile)
	if err != nil {
		return err
	}
	var detail *coverage.Detail
	if !ok {
		detail = &coverage.Detail{}
	} else {
		project, err := loadcov.LoadAll(loadcov.LoadAllOptions{
			Dir:      c.projectDir,
			Args:     c.args,
			Profiles: []string{c.coverageProfile},
			Ref:      git.COMMIT_WORKING,
			DiffBase: c.diffBase,
		})
		if err != nil {
			return err
		}
		summary := loadcov.ComputeCoverageSummary(project)
		base := summary[""]
		if base != nil {
			detail = &coverage.Detail{
				Total:       base.Total,
				Incrimental: base.Incrimental,
			}
		}
	}
	data, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	return os.WriteFile(c.cacheResultPath(), data, 0755)
}

func (c *controller) Clear() error {
	ok, err := fileExists(c.coverageProfile)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return os.RemoveAll(c.coverageProfile)
}

func fileExists(f string) (bool, error) {
	stat, err := os.Stat(f)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if stat.IsDir() {
		return false, fmt.Errorf("want file, but %s is a directory", f)
	}
	return true, nil
}

func (c *controller) cacheResultPath() string {
	return filepath.Join(c.covDir, "result.json")
}

func mapCovDir(tmpDir string, absDir string) string {
	list := strings.Split(absDir, string(filepath.Separator))
	for i, e := range list {
		list[i] = fileutil.CleanSpecial(e)
	}
	return filepath.Join(tmpDir, "xgo", "coverage", filepath.Join(list...))
}

func getTmpDir() string {
	if runtime.GOOS != "windows" {
		// try /tmp first for shorter path
		tmpDir := "/tmp"
		stat, statErr := os.Stat(tmpDir)
		if statErr == nil && stat.IsDir() {
			return tmpDir
		}
	}

	return os.TempDir()
}
