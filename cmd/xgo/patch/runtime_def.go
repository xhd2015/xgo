package patch

const RuntimeExtraDef = `
// xgo
func __xgo_getcurg() unsafe.Pointer
func __xgo_trap(interface{}, []interface{}, []interface{}) (func(), bool)
func __xgo_register_func(pkgPath string, fn interface{}, recvName string, argNames []string, resNames []string)
func __xgo_for_each_func(f func(pkgName string,funcName string, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string))
`

const RuntimeFuncNamePatch = `// workaround for go1.20, go1.21 will including this by go
func (md *moduledata) funcName(nameOff int32) string {
	if nameOff == 0 {
		return ""
	}
	return gostringnocopy(&md.funcnametab[nameOff])
}`

const NoderFiles = `	// auto gen
if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r io.Reader) {
		p := &noder{}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches) // errors are tracked via p.error
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
