package compiler_extra

import (
	"encoding/json"
	"os"
)

type Packages struct {
	Packages []*Package `json:"packages,omitempty"`
}

type Package struct {
	Path       string  `json:"path"`
	HasVarTrap bool    `json:"hasVarTrap"`
	Files      []*File `json:"files"`
}

type File struct {
	Name       string       `json:"name"`
	Funcs      []*Func      `json:"funcs,omitempty"`
	Interfaces []*Interface `json:"interfaces,omitempty"`
}

type Func struct {
	IdentityName string `json:"identityName"`
}

type Interface struct {
	Name string `json:"name"`
}

// mapping for efficient reading
type PackagesMapping struct {
	Packages map[string]*PackageMapping `json:"packages"`
}

type PackageMapping struct {
	HasVarTrap bool                    `json:"hasVarTrap"`
	Files      map[string]*FileMapping `json:"files"`
}

type FileMapping struct {
	Funcs      map[string]*Func      `json:"funcs,omitempty"`
	Interfaces map[string]*Interface `json:"interfaces,omitempty"`
}

func (c *Packages) BuildMapping() *PackagesMapping {
	if c == nil {
		return nil
	}
	mapping := make(map[string]*PackageMapping, len(c.Packages))
	for _, pkg := range c.Packages {
		files := make(map[string]*FileMapping, len(pkg.Files))
		for _, file := range pkg.Files {
			files[file.Name] = file.BuildMapping()
		}
		mapping[pkg.Path] = &PackageMapping{
			HasVarTrap: pkg.HasVarTrap,
			Files:      files,
		}
	}
	return &PackagesMapping{
		Packages: mapping,
	}
}

func (c *File) BuildMapping() *FileMapping {
	if c == nil {
		return nil
	}
	funcMapping := make(map[string]*Func, len(c.Funcs))
	for _, f := range c.Funcs {
		funcMapping[f.IdentityName] = f
	}
	interfaceMapping := make(map[string]*Interface, len(c.Interfaces))
	for _, i := range c.Interfaces {
		interfaceMapping[i.Name] = i
	}
	return &FileMapping{
		Funcs:      funcMapping,
		Interfaces: interfaceMapping,
	}
}

func ParseMapping(file string) (*PackagesMapping, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var mapping PackagesMapping
	err = json.Unmarshal(data, &mapping)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}
