package go_info

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type GoVersion struct {
	Major int // 1
	Minor int // 17
	Patch int // 5

	OS   string
	Arch string
}

func GetGoVersion() (*GoVersion, error) {
	// go version go1.17.5 darwin/amd64
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return nil, fmt.Errorf("cannot get go version")
	}
	outTrim := strings.TrimSuffix(string(out), "\n")
	return parseGoVersion(outTrim)
}

var gooVersionOnce sync.Once
var goVersion *GoVersion
var goVersionErr error

func GetGoVersionCached() (*GoVersion, error) {
	gooVersionOnce.Do(func() {
		goVersion, goVersionErr = GetGoVersion()
	})
	return goVersion, goVersionErr
}

const goVersionPrefix = "go version "

func parseGoVersion(s string) (*GoVersion, error) {
	if !strings.HasPrefix(s, goVersionPrefix) {
		return nil, fmt.Errorf("unrecognized version, expect prefix '%s': %s", goVersionPrefix, s)
	}
	s = s[len(goVersionPrefix):]
	if !strings.HasPrefix(s, "go") {
		return nil, fmt.Errorf("unrecognized version, expect pattern 'go1.x.y': %s", s)
	}
	s = s[len("go"):]

	spaceIdx := strings.Index(s, " ")
	if spaceIdx < 0 {
		return nil, fmt.Errorf("unrecognized version, expect space after 'go1.x.y': %s", s)
	}
	version := s[:spaceIdx]
	osArch := s[spaceIdx+1:]

	res := &GoVersion{}
	verList := strings.Split(version, ".")
	for i := 0; i < 3; i++ {
		if i < len(verList) {
			verInt, err := strconv.ParseInt(verList[i], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unrecognized version, expect number, found: %s", version)
			}
			switch i {
			case 0:
				res.Major = int(verInt)
			case 1:
				res.Minor = int(verInt)
			case 2:
				res.Patch = int(verInt)
			}
		}
	}
	slashIdx := strings.Index(osArch, "/")
	if slashIdx < 0 {
		return nil, fmt.Errorf("unrecognized version, expect os/arch: %s", osArch)
	}
	res.OS = osArch[:slashIdx]
	res.Arch = osArch[slashIdx+1:]
	return res, nil
}
