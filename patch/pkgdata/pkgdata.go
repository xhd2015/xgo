package pkgdata

import (
	"cmd/compile/internal/base"
	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type PackageData struct {
	Vars   map[string]bool
	Consts map[string]*ConstInfo
	Funcs  map[string]bool
}

type ConstInfo struct {
	Type    string // t=xx
	Untyped bool   // no 'e='(explicit) flag
}

var pkgDataMapping map[string]*PackageData

func GetPkgData(pkgPath string) *PackageData {
	data, ok := pkgDataMapping[pkgPath]
	if ok {
		return data
	}
	data, err := load(pkgPath)
	if err != nil {
		base.Errorf("load package data: %s %v", pkgPath, err)
		return nil
	}
	if pkgDataMapping == nil {
		pkgDataMapping = make(map[string]*PackageData, 1)
	}
	pkgDataMapping[pkgPath] = data
	return data
}
func WritePkgData(pkgPath string, pkgData *PackageData) error {
	file := getPkgDataFile(pkgPath)

	err := os.MkdirAll(filepath.Dir(file), 0755)
	if err != nil {
		return err
	}
	w, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer w.Close()

	err = writeConstSection(w, "[const]", pkgData.Consts)
	if err != nil {
		return err
	}
	writeSection(w, "[var]", pkgData.Vars)
	if err != nil {
		return err
	}
	writeSection(w, "[func]", pkgData.Funcs)
	if err != nil {
		return err
	}

	return nil
}
func writeSection(w io.Writer, section string, m map[string]bool) error {
	if len(m) == 0 {
		return nil
	}
	_, err := io.WriteString(w, section)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, "\n")
	if err != nil {
		return err
	}
	for k := range m {
		_, err := io.WriteString(w, k)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func writeConstSection(w io.Writer, section string, m map[string]*ConstInfo) error {
	if len(m) == 0 {
		return nil
	}
	_, err := io.WriteString(w, section)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, "\n")
	if err != nil {
		return err
	}
	for k, v := range m {
		_, err := io.WriteString(w, k)
		if err != nil {
			return err
		}
		err = writeConst(w, v)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\n")
		if err != nil {
			return err
		}
	}
	return nil
}
func writeConst(w io.Writer, info *ConstInfo) error {
	if !info.Untyped {
		_, err := io.WriteString(w, " ")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "e")
		if err != nil {
			return err
		}
	} else {
		// only untyped requires type info
		_, err := io.WriteString(w, " ")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "t=")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, info.Type)
		if err != nil {
			return err
		}
	}

	return nil
}

func load(pkgPath string) (*PackageData, error) {
	if xgo_ctxt.XgoCompilePkgDataDir == "" {
		return nil, fmt.Errorf("XGO_COMPILE_PKG_DATA_DIR not set")
	}
	file := getPkgDataFile(pkgPath)
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return parsePkgData(string(data))
}

type Section int

const (
	Section_Func  Section = 1
	Section_Var   Section = 2
	Section_Const Section = 3
)

func getPkgDataFile(pkgPath string) string {
	fsPath := pkgPath
	if filepath.Separator != '/' {
		split := strings.Split(pkgPath, "/")
		fsPath = filepath.Join(split...)
	}
	return filepath.Join(xgo_ctxt.XgoCompilePkgDataDir, fsPath, "__xgo_pkgdata__.txt")
}

// [func]
func parsePkgData(content string) (*PackageData, error) {
	lines := strings.Split(content, "\n")
	n := len(lines)

	p := &PackageData{}
	var section Section
	for i := 0; i < n; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch line {
		case "[func]":
			section = Section_Func
		case "[var]":
			section = Section_Var
		case "[const]":
			section = Section_Const
		default:
			if section == 0 {
				break
			}
			name := line
			idx := strings.Index(line, " ")
			if idx >= 0 {
				name = line[:idx]
			}
			extra := line[idx+1:]
			if name == "" {
				break
			}
			switch section {
			case Section_Func:
				if p.Funcs == nil {
					p.Funcs = make(map[string]bool, 1)
				}
				p.Funcs[name] = true
			case Section_Var:
				if p.Vars == nil {
					p.Vars = make(map[string]bool, 1)
				}
				p.Vars[name] = true
			case Section_Const:
				if p.Consts == nil {
					p.Consts = make(map[string]*ConstInfo, 1)
				}
				constInfo := &ConstInfo{Untyped: true}
				kvs := strings.Split(extra, " ")
				for _, v := range kvs {
					if v == "e" {
						constInfo.Untyped = false
						continue
					}
					if strings.HasPrefix(v, "t=") {
						constInfo.Type = v[len("t="):]
					}
				}
				p.Consts[name] = constInfo
			default:
				// ignore others
			}
		}
	}
	return p, nil
}
