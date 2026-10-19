package template

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

func InstantiateXgoTrap(xgoTrapTemplate string, goVersion *goinfo.GoVersion) (string, error) {
	// instrument xgo_trap.go
	xgoTrapContent, err := patch.RemoveBuildIgnore(xgoTrapTemplate)
	if err != nil {
		return "", fmt.Errorf("remove build ignore: %w", err)
	}
	// type _panic has a major upgrade from go1.21 to go1.22
	// go1.21 and before use pc, go1.22 and after use retpc
	retpc := "retpc"
	if goVersion.Major == 1 && goVersion.Minor <= 21 {
		retpc = "pc"
	}
	xgoTrapContent = strings.ReplaceAll(xgoTrapContent, "__RETPC__", retpc)
	xgoTrapContent = AppendGetFuncNameImpl(goVersion, xgoTrapContent)
	xgoTrapContent, err = AppendTrapSkipImpl(goVersion, xgoTrapContent)
	if err != nil {
		return "", err
	}

	return guardTrapEntries(xgoTrapContent)
}

// The template may come from an older xgo/runtime dependency, so install
// the safety checks here rather than relying on its version of the template.
func guardTrapEntries(content string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "xgo_trap.go", content, 0)
	if err != nil {
		return "", fmt.Errorf("parse xgo_trap.go: %w", err)
	}
	type insertion struct {
		offset int
		guard  string
	}
	var insertions []insertion
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Body == nil {
			continue
		}
		var ret string
		switch fn.Name.Name {
		case "XgoTrap":
			ret = "nil, false"
		case "XgoTrapVar", "XgoTrapVarPtr":
			ret = ""
		case "XgoTrapOn":
			ret = "false"
		default:
			continue
		}
		offset := fset.Position(fn.Body.Lbrace).Offset + 1
		guard := "\nif !__xgo_can_trap() { return " + ret + " }\n"
		if !strings.HasPrefix(content[offset:], guard) {
			insertions = append(insertions, insertion{offset, guard})
		}
	}
	// Insert backwards so positions still refer to the original source.
	for i := len(insertions) - 1; i >= 0; i-- {
		ins := insertions[i]
		content = content[:ins.offset] + ins.guard + content[ins.offset:]
	}
	if !strings.Contains(content, "func __xgo_can_trap()") {
		content += trapSafetyImpl
	}
	return content, nil
}

const trapSafetyImpl = `
// __xgo_can_trap must run before reading trap state or invoking user code.
// Runtime callbacks can run on g0 (even with m.curg set), on gsignal, or
// while holding runtime locks. They cannot safely allocate, defer or block.
//go:nosplit
func __xgo_can_trap() bool {
	gp := getg()
	return gp == gp.m.curg && gp.m.locks == 0
}
`

var runtimePkgClause = regexp.MustCompile(`(?m)^package runtime[ \t]*\r?\n`)

// AppendTrapSkipImpl adds XgoTrapOn and XgoSetTrapSkip to xgo_trap.go.
// They are generated here rather than kept in the template because the
// template comes from the xgo/runtime version of the project being built,
// which can predate them, while the instrumented code always links
// __xgo_trap_on_ to runtime.XgoTrapOn. Such a runtime never sets the skip,
// so its calls keep going through the trap. The runtime's atomic package
// moved in go1.23.
func AppendTrapSkipImpl(goVersion *goinfo.GoVersion, content string) (string, error) {
	if strings.Contains(content, trapOnDecl) {
		return content, nil
	}
	// instrumented code passes its flag to XgoTrapOn; a definition with any
	// other signature would leave those calls not compiling
	if strings.Contains(content, "func XgoTrapOn(") {
		return "", fmt.Errorf("xgo_trap.go: XgoTrapOn defined without %q", trapOnDecl)
	}
	atomicPkg := "internal/runtime/atomic"
	if goVersion.Major == 1 && goVersion.Minor <= 22 {
		atomicPkg = "runtime/internal/atomic"
	}
	loc := runtimePkgClause.FindStringIndex(content)
	if loc == nil {
		return "", fmt.Errorf("xgo_trap.go: package clause not found")
	}
	importDecl := fmt.Sprintf("\nimport __xgo_atomic %q\n", atomicPkg)
	return content[:loc[1]] + importDecl + content[loc[1]:] + trapSkipImpl, nil
}

const trapOnDecl = "func XgoTrapOn(flag *uint32) bool"

const trapSkipImpl = `
// __xgo_trap_skip is 1 while instrumented functions can skip calling
// XgoTrap, and building its arguments, because nothing could intercept the
// call. Only an xgo/runtime that turns it off again on first registration
// sets it. It is read on every instrumented call while registrations may
// run concurrently, hence atomic.
var __xgo_trap_skip uint32

// XgoTrapOn reports whether an instrumented function should call XgoTrap:
// when nothing makes every function skip it, or when something was
// registered for that function, which sets its flag.
func XgoTrapOn(flag *uint32) bool {
	return __xgo_atomic.Load(&__xgo_trap_skip) == 0 || __xgo_atomic.Load(flag) != 0
}

func XgoSetTrapSkip(skip bool) {
	var v uint32
	if skip {
		v = 1
	}
	__xgo_atomic.Store(&__xgo_trap_skip, v)
}
`
