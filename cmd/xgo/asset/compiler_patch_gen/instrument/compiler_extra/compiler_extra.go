package compiler_extra

import (
	"encoding/json"
	"os"
)

type Packages struct {
	Packages []*Package `json:"packages,omitempty"`
}

type Package struct {
	Path       string  `json:"path,omitempty"`
	HasVarTrap bool    `json:"hasVarTrap,omitempty"`
	Files      []*File `json:"files,omitempty"`
}

type File struct {
	Name       string       `json:"name,omitempty"`
	Funcs      []*Func      `json:"funcs,omitempty"`
	Interfaces []*Interface `json:"interfaces,omitempty"`
}

type Func struct {
	IdentityName string `json:"identityName,omitempty"`
}

type Interface struct {
	Name string `json:"name,omitempty"`
}

// mapping for efficient reading
// for large project, the mapping file can be
// as large as 13MB
type PackagesMapping struct {
	Packages map[string]*PackageMapping `json:"packages,omitempty"`
}

type PackageMapping struct {
	HasVarTrap bool                    `json:"hasVarTrap,omitempty"`
	Files      map[string]*FileMapping `json:"files,omitempty"`
}

type FileMapping struct {
	Funcs      map[string]*FuncMapping      `json:"funcs,omitempty"`
	Interfaces map[string]*InterfaceMapping `json:"interfaces,omitempty"`
}
type FuncMapping struct{}
type InterfaceMapping struct{}

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
	funcMapping := make(map[string]*FuncMapping, len(c.Funcs))
	for _, f := range c.Funcs {
		funcMapping[f.IdentityName] = &FuncMapping{}
	}
	interfaceMapping := make(map[string]*InterfaceMapping, len(c.Interfaces))
	for _, i := range c.Interfaces {
		interfaceMapping[i.Name] = &InterfaceMapping{}
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
