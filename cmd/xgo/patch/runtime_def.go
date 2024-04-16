package patch

const RuntimeProcPatch = `__xgo_is_init_finished = true
for _, fn := range __xgo_on_init_finished_callbacks {
	fn()
}
__xgo_on_init_finished_callbacks = nil
`

const RuntimeProcGoroutineCreatedPatch = `for _, fn := range __xgo_on_gonewproc_callbacks {
	fn(uintptr(unsafe.Pointer(xgo_newg)))
}
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

const RuntimeGetFuncName_Go117_120 = `
func __xgo_get_pc_name_impl(pc uintptr) string {
	return FuncForPC(pc).Name()
}
`

// start with go1.21, the runtime.FuncForPC(pc).Name()
// was wrapped in funcNameForPrint(...), we unwrap it
// NOTE: when upgrading to go1.23, should check
// the implementation again
const RuntimeGetFuncName_Go121 = `
func __xgo_get_pc_name_impl(pc uintptr) string {
	return FuncForPC(pc).__xgo_no_print_name()
}

func (f *Func) __xgo_no_print_name() string {
	if f == nil {
		return ""
	}
	fn := f.raw()
	if fn.isInlined() { // inlined version
		fi := (*funcinl)(unsafe.Pointer(fn))
		return fi.name
	}
	return funcname(f.funcInfo())
}
`

// FuncForPC(pc).Name() just works fine, there is no
// string split
const RuntimeGetFuncName_Go120_Unused = `
func __xgo_get_pc_name_impl(pc uintptr) string {
	fn := findfunc(pc)
	return fn.datap.funcName(fn.nameOff)
}
// workaround for go1.20, go1.21 will including this by go
func (md *moduledata) funcName(nameOff int32) string {
	if nameOff == 0 {
		return ""
	}
	return gostringnocopy(&md.funcnametab[nameOff])
}
`

const NoderFiles_1_17 = `	// auto gen
if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r io.Reader) *syntax.File {
		p := &noder{
			err: make(chan syntax.Error),
		}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches)
		if err != nil {
			e := err.(syntax.Error)
			p.error(e)
			return nil
		}
		p.file = file
		noders = append(noders, p)

		// move to head
		n := len(noders)
		for i:=n-1;i>0;i--{
			noders[i]=noders[i-1]
		}
		noders[0]=p

		return file
	})
}
`

const NoderFiles_1_20 = `	// auto gen
if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r io.Reader) *syntax.File {
		p := &noder{}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches)
		if err != nil {
			e := err.(syntax.Error)
			base.ErrorfAt(p.makeXPos(e.Pos), "%s", e.Msg)
			return nil
		}
		p.file = file
		noders = append(noders, p)

		// move to head
		n := len(noders)
		for i:=n-1;i>0;i--{
			noders[i]=noders[i-1]
		}
		noders[0]=p

		return file
	})
}
`

const NoderFiles_1_21 = `	// auto gen
if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r io.Reader) *syntax.File {
		p := &noder{}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches)
		if err != nil {
			e := err.(syntax.Error)
			base.ErrorfAt(m.makeXPos(e.Pos), 0,"%s", e.Msg)
			return nil
		}
		p.file = file
		noders = append(noders, p)

		// move to head
		n := len(noders)
		for i:=n-1;i>0;i--{
			noders[i]=noders[i-1]
		}
		noders[0]=p

		return file
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

const NodesGen = `
func (n *node) SetPos(p Pos) {
	n.pos = p
}
`

const Nodes_Inspect_117 = `
// Walk stops when f returns true, so invert it here
func Inspect(root Node, f func(Node) bool) {
	Walk(root, func(n Node) bool {
		return !f(n)
	})
}
`
