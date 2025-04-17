package goinfo

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type GoVersion struct {
	Major int // 1
	Minor int // 17
	Patch int // 5

	OS   string
	Arch string
}

func (c *GoVersion) String() string {
	return fmt.Sprintf("go version go%d.%d.%d %s/%s", c.Major, c.Minor, c.Patch, c.OS, c.Arch)
}

const goVersionPrefix = "go version "

func GetGoVersionOutput(goBinary string) (string, error) {
	out, err := exec.Command(goBinary, "version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

func GetGorootVersion(goroot string) (*GoVersion, error) {
	goBinary := filepath.Join(goroot, "bin", "go")
	version, err := GetGoVersionOutput(goBinary)
	if err != nil {
		return nil, err
	}
	return ParseGoVersion(version)
}

func ParseGoVersion(s string) (*GoVersion, error) {
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

	res, err := ParseGoVersionNumber(version)
	if err != nil {
		return nil, err
	}

	slashIdx := strings.Index(osArch, "/")
	if slashIdx < 0 {
		return nil, fmt.Errorf("unrecognized version, expect os/arch: %s", osArch)
	}
	res.OS = osArch[:slashIdx]
	res.Arch = osArch[slashIdx+1:]
	return res, nil
}

// ParseGoVersionNumber parses
// example input: 1.23rc1
func ParseGoVersionNumber(version string) (*GoVersion, error) {
	res := &GoVersion{}
	verList := strings.Split(version, ".")
	for i := 0; i < 3; i++ {
		if i < len(verList) {
			num := verList[i]
			if i == 1 {
				// 1.23rc1
				idx := strings.IndexFunc(num, func(r rune) bool {
					return r < '0' || r > '9'
				})
				if idx > 0 {
					num = num[:idx]
				}
			}
			verInt, err := strconv.ParseInt(num, 10, 64)
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
	return res, nil
}
