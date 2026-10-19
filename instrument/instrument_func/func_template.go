package instrument_func

import (
	"fmt"
	"strings"
)

// funcInfo, recvPtr, paramPtrs, resultPtrs
const trapTemplate = "__xgo_post_%d, __xgo_stop_%d := __xgo_trap_%d(%s,%s,[]interface{}{%s},[]interface{}{%s});if __xgo_post_%d!=nil { defer __xgo_post_%d(); };if __xgo_stop_%d { return; };"

// guardedTrap builds a trap statement that only calls the trap while
// __xgo_trap_on_<fileIndex>() reports true, and passes it pointers to copies
// of the receiver, parameters and results instead of the variables
// themselves. Taking the address of a variable that is passed to the trap
// moves it to the heap for the whole function, so trapTemplate costs
// several allocations on every call even when nothing intercepts it; with
// copies, only the guarded branch allocates. The copies are written back
// after the trap returns (the parameters and results the interceptor may
// have changed are visible to the body), and synced in both directions
// around post by helper.
//
// helper is a top-level function declared next to the instrumented one: a
// function literal would shift the numbering of the closures in the body
// (F.func1 becoming F.func2), and pointers passed to a function that does
// not keep them do not move the variables to the heap. The returned helper
// declaration is a single line, so it does not shift line numbers either.
//
// vars lists the receiver (if any), parameters and results, in that order.
func guardedTrap(line int, fileIndex int, helper string, funcInfo string, recv *guardVar, params []*guardVar, results []*guardVar) (stmt string, helperDecl string) {
	var vars []*guardVar
	if recv != nil {
		vars = append(vars, recv)
	}
	vars = append(vars, params...)
	nIn := len(vars)
	vars = append(vars, results...)

	names := make([]string, len(vars))
	copies := make([]string, len(vars))
	for i, v := range vars {
		names[i] = v.name
		copies[i] = fmt.Sprintf("__xgo_c_%d_%d", line, i)
	}
	addrs := func(c []string) string {
		a := make([]string, len(c))
		for i, name := range c {
			a[i] = "&" + name
		}
		return strings.Join(a, ",")
	}
	assign := func(dst []string, src []string) string {
		if len(dst) == 0 {
			return ""
		}
		return strings.Join(dst, ",") + "=" + strings.Join(src, ",") + ";"
	}

	recvAddr := "nil"
	paramCopies := copies[:nIn]
	if recv != nil {
		recvAddr = "&" + copies[0]
		paramCopies = copies[1:nIn]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "if __xgo_trap_on_%d(&%s) {", fileIndex, trapOnVar(funcInfo))
	if len(vars) > 0 {
		fmt.Fprintf(&b, "%s:=%s;", strings.Join(copies, ","), strings.Join(names, ","))
	}
	fmt.Fprintf(&b, "__xgo_post_%d, __xgo_stop_%d := __xgo_trap_%d(%s,%s,[]interface{}{%s},[]interface{}{%s});", line, line, fileIndex, funcInfo, recvAddr, addrs(paramCopies), addrs(copies[nIn:]))
	b.WriteString(assign(names, copies))
	if len(vars) == 0 {
		fmt.Fprintf(&b, "if __xgo_post_%d!=nil { defer __xgo_post_%d() };", line, line)
	} else {
		pairs := make([]string, 0, 2*len(vars))
		for i := range vars {
			pairs = append(pairs, "&"+names[i], "&"+copies[i])
		}
		fmt.Fprintf(&b, "if __xgo_post_%d!=nil { defer %s(__xgo_post_%d,%s) };", line, helper, line, strings.Join(pairs, ","))
	}
	fmt.Fprintf(&b, "if __xgo_stop_%d { return; };", line)
	b.WriteString("};")
	if len(vars) == 0 {
		return b.String(), ""
	}

	hParams := make([]string, 0, 2*len(vars))
	hVars := make([]string, len(vars))
	hCopies := make([]string, len(vars))
	for i, v := range vars {
		hVars[i] = fmt.Sprintf("*__xgo_v%d", i)
		hCopies[i] = fmt.Sprintf("*__xgo_c%d", i)
		hParams = append(hParams, fmt.Sprintf("__xgo_v%d,__xgo_c%d *%s", i, i, v.typ))
	}
	var h strings.Builder
	fmt.Fprintf(&h, ";func %s(post func(),%s) {%spost();%s}", helper, strings.Join(hParams, ","), assign(hCopies, hVars), assign(hVars[nIn:], hCopies[nIn:]))
	return b.String(), h.String()
}

// guardedDirectTrap is trapTemplate behind the __xgo_trap_on_<fileIndex>()
// guard, for functions whose signature guardedTrap cannot spell outside of
// them (generic functions and methods). It still skips the trap call, but
// the variables whose addresses it takes live on the heap.
func guardedDirectTrap(line int, fileIndex int, funcInfo string, recvAddr string, paramAddrs []string, resultAddrs []string) string {
	return fmt.Sprintf("if __xgo_trap_on_%d(&%s) {", fileIndex, trapOnVar(funcInfo)) + fmt.Sprintf(trapTemplate,
		line, line,
		fileIndex,
		funcInfo,
		recvAddr,
		strings.Join(paramAddrs, ","),
		strings.Join(resultAddrs, ","),
		line, line, line,
	) + "};"
}

// guardVar is a receiver, parameter or result of an instrumented function,
// with its type as source text.
type guardVar struct {
	name string
	typ  string
}

// trapOnVar is the name of the flag that the guard of the function with
// info variable funcInfo passes to __xgo_trap_on_. The runtime sets it
// when something is registered for that function.
func trapOnVar(funcInfo string) string {
	return funcInfo + "_on"
}
