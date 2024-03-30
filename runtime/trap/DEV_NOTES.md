# Problems
Infinite loop: trap into interceptor, should not be allowed

findfunc uses linear search, which might be slow when there are many modules, maybe a cache would help?


# TODO
Is there a way to enumerate all functions?

What is a moduledata in the runtime's perspective, is it a go package or a go module(not likely)?


Maybe go link is better than init?Given that init needs some extra effort to hack.

# Knowledge
runtime.Callers(skip, []pc) returns a slice of pcs of current stack

runtime.CallersFrame() returns an iterator over a slice of pcs which can be used to retrieve all frames as needed.


## the runtime.moduledata

minpc,maxpc   ---> used to search pc
inittasks -> a list of init tasks
modulename ---> 

ftab  --> a list of offset and entry info of all funcs, offset are to be used in pclntable
pclntable --> pclntable[funcOff] is type of _func
     example:
			f1 := funcInfo{(*_func)(unsafe.Pointer(&datap.pclntable[datap.ftab[i].funcoff])), datap}

ptab -> a list of exported functions

itablinks []*itab  -> interface,type table

## funcnametab
this section are all function names separated by `\x00`.
```go
func printFuncNames(funcnametab []byte) {
	n := len(funcnametab)
	last := -1
	for i := 0; i < n; i++ {
		if funcnametab[i] == '\x00' {
			println(string(funcnametab[last+1 : i]))
			last = i
		}
	}
}
```
Would print about 2931 names, like:
```
go:buildid
...
slices.Grow[go.shape.[]uint8,go.shape.uint8]
...
encoding/json.appendString[go.shape.string]
slices.SortFunc[go.shape.[]encoding/json.reflectWithString,go.shape.struct { encoding/json.v reflect.Value; encoding/json.ks string }]
type:.eq.encoding/json.reflectWithString
type:.eq.struct { encoding/json.ptr interface {}; encoding/json.len int }
type:.eq.go.shape.struct { encoding/json.v reflect.Value; encoding/json.ks string }
github.com/xhd2015/xgo/runtime/pkg.Hello
github.com/xhd2015/xgo/runtime/pkg.Mass.Print
github.com/xhd2015/xgo/runtime/pkg.(*Person).Greet
github.com/xhd2015/xgo/runtime/pkg.Hello.func1
main.init.0
main.main
main.testArgs
main.num.add
```

In brief, it contains all functions compiled/linked into the binary, so that gives us a chance to list all functions.

Note, there are forms like `go:buildid`,`slices.Grow[go.shape.[]uint8,go.shape.uint8]`, the `[...]` denotes instantiated generic params.

## runtime._func
> // Layout of in-memory per-function information prepared by linker<br>
  // See https://golang.org/s/go12symtab. <br>
  type _func struct { <br>
  ...


`pc(entryOff) = datap.text + entryOff`
```go
func (f funcInfo) entry() uintptr {
	return f.datap.textAddr(f.entryOff)
}
```

## reflect.Func
```
// Non-nil func value points at data block.
// First word of data block is actual code.
```

NOTE: cannot take address of a function
```
p := &testReflect
ERROR: invalid operation: cannot take address of testReflect (value of type func())
```
A function symbol is itself a pointer to the function entry.

```go
// f itself is a named variable in some place, its type is *byte
var v interface{} = f  ----> v.word = &f
reflect.ValueOf(f)  ---> 
```
A reflect.ValueOf(v) is just a wrapper around interface{}


Test pc meaning:
```go
func main() {
	testReflect()
	fnWord := getReflectWord(testReflect)
	fmt.Printf("testReflect word: %x\n", fnWord)
	fnAddrPtr := (*unsafe.Pointer)((unsafe.Pointer)(fnWord))
	fmt.Printf("testReflect word target: %x\n", *fnAddrPtr)
	fmt.Println(testReflect)
}
func testReflect() {
	pc := runtime.Getcallerpc()
	entryPC := runtime.GetcallerFuncPC()

	fmt.Printf("testReflect caller pc: %x\n", pc)
	fmt.Printf("testReflect caller entry pc: %x\n", entryPC)
}

```

Output:
```sh
testReflect caller pc: c6423b5
testReflect caller entry pc: c642300
testReflect word: c678298
testReflect word target: c642300
0xc642300
```

Found that **entryPC is the same thing with function symbol**, this is a very important observation.

Explanation:
   a function symbol is entry to the function body, function types are either inserted by compiler statically or carried by interface dynamically. So a function symbol is considered *byte=PC, pointer to a readonly part.

   an interface is a {type,word}, the ptr itself is allocated on heap, it has type *PC, i.e.  {type:funcType, word: *PC}

## getReflectWord
Get address of an interface
```go
func getReflectWord(i interface{}) uintptr {
	type IHeader struct {
		typ  uintptr
		word uintptr
	}

	return (*IHeader)(unsafe.Pointer(&i)).word
}
```

## How to list all functions at runtime?
```go
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
```

Output:
```sh
ftab len: 2933
ftab:0 0xe06e000 
ftab:1 0xe06e080 internal/abi.(*RegArgs).IntRegArgAddr
...
ftab:48 0xe06f5c0 type:.eq.internal/abi.UncommonType
ftab:49 0xe06f600 type:.eq.internal/abi.RegArgs
...
ftab:2930 0xe153f60 main.testArgs
ftab:2931 0xe154300 main.num.add
ftab:2932 0xe15465f lBreak
```

NOTE: there are some names starting with prefix `type:`.

## How to get runtime type of a function
What is a type? Look at the interface{} structure:
```go

```
Use `runtime.resolveTypeOff`
```
type moduledata{
	// ...
    types, etypes         uintptr
	// ...
}
types and etypes are the range of type data

```

## How to construct an interface{} for a func using pc?
```
```

## How to invoke a function
First, construct an interface with `type` set to func type, `word` set pointer to pc.



## What about parameter names

## symtab
[../../runtime/symtab.go](../../runtime/symtab.go)

## How to construct a reflect.Value from pc?
Through my investigation, there is no type info from a PC value.
Types are inserted at compile time by compiler.

A workaround: when calling `__x_trap()`, carry the function symbol with itself.

And for registration and invoking purepose, we make the program register the types automatically.

```go
func init(){
	// building a PC -> type mapping
	registerFunc(A)
	registerFunc(B)
}

func A(){
 ...
}
type T struct{}
func (c *T) A(){
 ...
}

// T.A
// *T.A
```


# Empty interface vs interface with methods
The reflect implementation:

```go
// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  *abi.Type
	word unsafe.Pointer
}

// nonEmptyInterface is the header for an interface value with methods.
type nonEmptyInterface struct {
	// see ../runtime/iface.go:/Itab
	itab *struct {
		ityp *abi.Type // static interface type
		typ  *abi.Type // dynamic concrete type
		hash uint32    // copy of typ.hash
		_    [4]byte
		fun  [100000]unsafe.Pointer // method table
	}
	word unsafe.Pointer
}

```

## funcType
```go
// returns a function of
reflect.FuncOf = func(in, out []Type, variadic bool) Type
```

`funcType`
```go
// funcType represents a function type.
//
// A *rtype for each in and out parameter is stored in an array that
// directly follows the funcType (and possibly its uncommonType). So
// a function type with one method, one input, and one output is:
//
//	struct {
//		funcType
//		uncommonType
//		[2]*rtype    // [0] is in, [1] is out
//	}
type funcType = abi.FuncType
```

# type at runtime
```go
// reflectOffs holds type offsets defined at run time by the reflect package.
//
// When a type is defined at run time, its *rtype data lives on the heap.
// There are a wide range of possible addresses the heap may use, that
// may not be representable as a 32-bit offset. Moreover the GC may
// one day start moving heap memory, in which case there is no stable
// offset that can be defined.
//
// To provide stable offsets, we add pin *rtype objects in a global map
// and treat the offset as an identifier. We use negative offsets that
// do not overlap with any compile-time module offsets.
//
// Entries are created by reflect.addReflectOff.
var reflectOffs struct {
	lock mutex
	next int32
	m    map[int32]unsafe.Pointer
	minv map[unsafe.Pointer]int32
}
```

# FuncData
Including args info.

see `go/src/internal/abi/symtab.go`

```go
FUNCDATA_ArgsPointerMaps    = 0
FUNCDATA_LocalsPointerMaps  = 1
FUNCDATA_StackObjects       = 2
FUNCDATA_InlTree            = 3
FUNCDATA_OpenCodedDeferInfo = 4
FUNCDATA_ArgInfo            = 5
FUNCDATA_ArgLiveInfo        = 6
FUNCDATA_WrapInfo           = 7
```
# PCDATA
UnsafePoint is for gc?
```go
PCDATA_UnsafePoint   = 0
PCDATA_StackMapIndex = 1
PCDATA_InlTreeIndex  = 2
PCDATA_ArgLiveIndex  = 3
```

# IR
## insert a function call
```go
func addPrint(){
	for _, fn := range typecheck.Target.Funcs {
		callPrint := ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, typecheck.LookupRuntime("printstring"), []ir.Node{
			ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("hello init\n")),
		})
	    callPrint = typecheck.Expr(callPrint)
		fn.Body.Prepend(callPrint)
	}
}
```

## add an init function
- use `types.LocalPkg` to create a pkg scope symbol
- use `ir.NewFunc` to create the function, and create its body
- use `typecheck.Stmts` to typecheck it's body(which will probably normalize expr to statement if needed)
- append to `typecheck.Target.Inits` and `typecheck.Target.Funcs`
```go
func addInit(){
	// init names are usually init.0, init.1, ...
	sym,exists := types.LocalPkg.LookupOK(fmt.Sprintf("init.%d", len(typecheck.Target.Inits)))
	if exists {
		panic(fmt.Errorf("init name error"))
	}
	regFuncs := ir.NewFunc(base.AutogeneratedPos, base.AutogeneratedPos, sym, types.NewSignature(nil, nil, nil))

	regFuncs.Body = []ir.Node{
		ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, typecheck.LookupRuntime("printstring"), []ir.Node{
			ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("hello init\n")),
		}),
	}

	// this typecheck is required
	// to make subsequent steps work
	typecheck.Stmts(regFuncs.Body)

	typecheck.Target.Inits = append(typecheck.Target.Inits, regFuncs)
	typecheck.Target.Funcs = append(typecheck.Target.Funcs, regFuncs)
}
```

# About Method Receiver
```go

// this function checks if the given
// `recvPtr` has the same value compared
// to the given `methodValue`.
// The `methodValue` should be passed as
// `file.Write“.
// Deprecated: left here only for reference purepose
func isSameBoundMethod(recvPtr interface{}, methodValue interface{}) bool {
	// can also be a constant
	// size := unsafe.Sizeof(*(*large)(nil))
	size := reflect.TypeOf(recvPtr).Elem().Size()
	type _intfRecv struct {
		_    uintptr // type word
		data *byte   // data word
	}

	a := (*_intfRecv)(unsafe.Pointer(&recvPtr))
	type _methodValue struct {
		_    uintptr // pc
		recv byte
	}
	type _intf struct {
		_    uintptr // type word
		data *_methodValue
	}
	ppb := (*_intf)(unsafe.Pointer(&methodValue))
	pb := *ppb
	b := unsafe.Pointer(&pb.data.recv)

	return __xgo_link_mem_equal(unsafe.Pointer(a.data), b, size)
}
```

# About function wrapper
The following code does not work, because there are some wrapper around functions, so PC's are not always the same.
```go
// check if the calling func is an interceptor, if so, skip
// UPDATE: don't do manual check
for i := 0; i < n; i++ {
	if interceptors[i].Pre == nil {
		continue
	}
	ipc := (**uintptr)(unsafe.Pointer(&interceptors[i].Pre))
	pcName := runtime.FuncForPC(**ipc).Name()
	_ = pcName
	if **ipc == pc {
		return nil, false
	}
}
```

# A thinking on type safe mock
```go
OnFunc(hello, func(a string) {
	if a == "" {
		CallOld()
		// be able to call old function
	}

	return
})


func OnFunc[T any](fn T, v T) {

}

```