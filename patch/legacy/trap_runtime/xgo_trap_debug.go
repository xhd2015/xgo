// This file should be inserted into runtime to provide hack points

//go:build ignore

package runtime

import (
	"unsafe"
)

// important thing here:
//
//		get arg types and argNames
//	 get names
//
// two ways to
//
//	by function name
//
// it needs defer
//
//	if something{
//	    next()
//	    myAction()
//	}

// other interesting things:
//  iterate all functions

func TestModuleDataGetType_Requires_Xgo() (start uintptr, end uintptr) {
	for p := &firstmoduledata; p != nil; p = p.next {
		return p.types, p.etypes
	}
	throw("shit")
	return 0, 0
}

func TestModuleDataFindTypeLink_Requires_Xgo(offset int32) int {
	for p := &firstmoduledata; p != nil; p = p.next {
		for i, tl := range p.typelinks {
			if offset == tl {
				return i
			}
		}
	}
	return -1
}
func TestModuleData_Requires_Xgo() {
	for p := &firstmoduledata; p != nil; p = p.next {
		println("module:", p.modulename)
		println("hasmain:", p.hasmain)
		printFuncNames(p.funcnametab)
		printPTab(p, p.ptab)
		printHeader(p.pcHeader)
		printFTab(p, p.ftab)
	}
}
func printHeader(p *pcHeader) {
	println("magic:", p.magic) // 0xfffffff1, i.e. 4294967281
	println("nfunc:", p.nfunc)
	println("nfiles:", p.nfiles)
}

func printPTab(m *moduledata, ptab []ptabEntry) {
	println("ptab:", len(ptab))
	for i, p := range ptab {
		println("ptab:", i, m.funcName(int32(p.name)))
	}
}
func printFuncNames(funcnametab []byte) {
	n := len(funcnametab)
	last := -1
	num := 0
	for i := 0; i < n; i++ {
		if funcnametab[i] == '\x00' {
			// println(string(funcnametab[last+1 : i]))
			last = i
			num++
		}
	}
	_ = last
	println("funcnametab len:", num)
}
func printFTab(m *moduledata, ftab []functab) {
	println("ftab len:", len(ftab))
	for i, f := range ftab {
		// funcoff -> offset to function info, like name
		// pc,_ := m.textOff(uintptr(f.entryoff))
		pc := m.textAddr(f.entryoff)
		fnInfo := funcInfo{(*_func)(unsafe.Pointer(&m.pclntable[f.funcoff])), m}
		print("ftab:", i)
		printsp()
		printhex(uint64(pc))
		printsp()
		println(m.funcName(fnInfo.nameOff))
	}
}

// all are ptrs
func __x_trap2(recv interface{}, args []interface{}, results []interface{}) bool {
	println("recv: ", recv)
	println("args: ", args)
	println("results: ", results)

	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			println("args[0]: ", s)
		}
		if s, ok := args[0].(*string); ok {
			println("args[0]: ", *s)
			*s = *s + "_trap"
			println("modified")
		}
	}
	if len(results) > 0 {
		if i, ok := results[0].(*int); ok {
			*i = 20
			return true
		}
	}

	// TODO: how to reveal args?

	// reveal func
	pc := getcallerpc()
	fn := findfunc(pc)
	// TODO: what about inlined func?
	funcName := fn.datap.funcName(fn.nameOff)

	println("caller func:", funcName)
	return false
}
func __x_trap_trace(s string) {
	print("entered trap\n")
	goroutineheader := func(gp *g) {
		println("goroutine ", gp.goid)
	}
	traceback2 := func(u *unwinder, showRuntime bool, skip, max int) (n, lastN int) {
		// commitFrame commits to a logical frame and returns whether this frame
		// should be printed and whether iteration should stop.
		commitFrame := func() (pr, stop bool) {
			if skip == 0 && max == 0 {
				// Stop
				return false, true
			}
			n++
			lastN++
			if skip > 0 {
				// Skip
				skip--
				return false, false
			}
			// Print
			max--
			return true, false
		}

		gp := u.g.ptr()
		level, _, _ := gotraceback()
		for ; u.valid(); u.next() {
			lastN = 0
			f := u.frame.fn
			for iu, uf := newInlineUnwinder(f, u.symPC()); uf.valid(); uf = iu.next(uf) {
				sf := iu.srcFunc(uf)
				callee := u.calleeFuncID
				u.calleeFuncID = sf.funcID
				if !(showRuntime || showframe(sf, gp, n == 0, callee)) {
					continue
				}

				if pr, stop := commitFrame(); stop {
					return
				} else if !pr {
					continue
				}

				name := sf.name()
				file, line := iu.fileLine(uf)
				// Print during crash.
				//	main(0x1, 0x2, 0x3)
				//		/home/rsc/go/src/runtime/x.go:23 +0xf
				//
				printFuncName(name)
				print("(")
				if iu.isInlined(uf) {
					print("...")
				} else {
					argp := unsafe.Pointer(u.frame.argp)
					printArgs(f, argp, u.symPC())
				}
				print(")\n")
				print("\t", file, ":", line)
				if !iu.isInlined(uf) {
					if u.frame.pc > f.entry() {
						print(" +", hex(u.frame.pc-f.entry()))
					}
					if gp.m != nil && gp.m.throwing >= throwTypeRuntime && gp == gp.m.curg || level >= 2 {
						print(" fp=", hex(u.frame.fp), " sp=", hex(u.frame.sp), " pc=", hex(u.frame.pc))
					}
				}
				print("\n")
			}
		}
		return
	}
	traceback1 := func(pc, sp, lr uintptr, gp *g, flags unwindFlags) {
		flags |= unwindPrintErrors
		var u unwinder
		tracebackWithRuntime := func(showRuntime bool) int {
			const maxInt int = 0x7fffffff
			u.initAt(pc, sp, lr, gp, flags)
			n, lastN := traceback2(&u, showRuntime, 0, tracebackInnerFrames)
			if n < tracebackInnerFrames {
				// We printed the whole stack.
				return n
			}
			_ = lastN
			return n
		}
		tracebackWithRuntime(false)
	}
	traceback := func(pc, sp, lr uintptr, gp *g) {
		traceback1(pc, sp, lr, gp, 0)
	}

	gp := getg()
	println("g goid:", gp.goid)
	println("curg goid:", gp.m.curg.goid)

	sp := getcallersp()
	pc := getcallerpc()
	systemstack(func() {
		println("system stack")
		g0 := getg()
		println("g0 goid:", g0.goid)
		// Force traceback=1 to override GOTRACEBACK setting,
		// so that Stack's results are consistent.
		// GOTRACEBACK is only about crash dumps.
		g0.m.traceback = 1
		// g0.writebuf = buf[0:0:0]
		println("before header")
		goroutineheader(gp)
		traceback(pc, sp, 0, gp)

		g0.m.traceback = 0
		g0.writebuf = nil
	})
}
