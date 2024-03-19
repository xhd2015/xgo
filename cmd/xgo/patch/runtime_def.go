package patch

const RuntimeProcPatch = `__xgo_is_init_finished = true
for _, fn := range __xgo_on_init_finished_callbacks {
	fn()
}
__xgo_on_init_finished_callbacks = nil
`

// added after goroutine exit1
const RuntimeProcGoroutineExitPatch = `for _, fn := range __xgo_on_goexits {
	fn()
}`

const TestingCallbackDeclarations = `func __xgo_link_get_test_starts() []interface{}{
	// link by compiler
	return nil
}
`

const TestingStart = `for _,__xgo_on_test_start:=range __xgo_link_get_test_starts(){
	(__xgo_on_test_start.(func(*T,func(*T))))(t,fn)
}
`

const RuntimeFuncNamePatch = ""

// Not used because now we pass pkg name, func name as standalone strings
const RuntimeFuncNamePatch_Not_Used = `// workaround for go1.20, go1.21 will including this by go
func (md *moduledata) funcName(nameOff int32) string {
	if nameOff == 0 {
		return ""
	}
	return gostringnocopy(&md.funcnametab[nameOff])
}`

const NoderFiles_1_17 = `	// auto gen
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
			p.error(e)
			return
		}
		p.file = file
		noders = append(noders, p)
	})
}
`

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

const NoderFiles_1_21 = `	// auto gen
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

const GenericTrapForGo118And119 = `// for all generic functions, add trap before them.
// NOTE this is a workaround for go1.18 and go1.19,
// because capturing generic variable after instantiation
// needs to construct a complex IR. However,if do that 
// before BuildInstantiations(), only simple name and type needed.
//
// TODO: change go1.20 and above to use this strategy, because 
// it has lower brain overhead
if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
	for _, decl := range g.target.Decls {
		fnDecl, ok := decl.(*ir.Func)
		if !ok {
			continue
		}
		if fnDecl.Type().NumTParams() == 0 {
			continue
		}

		_, ok = xgo_patch.CanInsertTrapOrLink(fnDecl)
		if !ok {
			continue
		}
		xgo_patch.InsertTrapForFunc(fnDecl, true)
	}
}
`
