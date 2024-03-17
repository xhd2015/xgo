package patch

const RuntimeFuncNamePatch = ""

// Not used because now we pass pkg name, func name as standalone strings
const RuntimeFuncNamePatch_Not_Used = `// workaround for go1.20, go1.21 will including this by go
func (md *moduledata) funcName(nameOff int32) string {
	if nameOff == 0 {
		return ""
	}
	return gostringnocopy(&md.funcnametab[nameOff])
}`

const NoderFiles_1_20 = `	// auto gen
if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r io.Reader) {
		p := &noder{}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches)
		if err != nil {
			e := err.(syntax.Error)
			base.ErrorfAt(p.makeXPos(e.Pos), "%s", e.Msg)
			return
		}
		p.file = file
		noders = append(noders, p)
	})
}
`

const NoderFiles_1_22 = `	// auto gen
if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r io.Reader) {
		p := &noder{}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches)
		if err != nil {
			e := err.(syntax.Error)
			base.ErrorfAt(m.makeXPos(e.Pos), 0,"%s", e.Msg)
			return
		}
		p.file = file
		noders = append(noders, p)
	})
}
`
