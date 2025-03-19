package trace_runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/xhd2015/xgo/runtime/trace/constants"
)

func nothing() {}

func SetupTrap() {
	runtime.XgoSetTrap(trap)
}

// type XgoField struct {
// 	Name string
// 	Ptr  interface{}
// }

func trap(recv runtime.XgoField, args []runtime.XgoField, results []runtime.XgoField) func() {
	// skip 2: <user func> -> runtime.XgoTrap -> trap
	const SKIP = 2

	var pcs [1]uintptr
	runtime.Callers(SKIP+1, pcs[:])
	pc := pcs[0]
	funcInfo := runtime.FuncForPC(pc)

	fnName := funcInfo.Name()

	stack := GetStack()
	var isStart bool
	if stack == nil {
		if fnName != constants.START_XGO_TRACE {
			return nothing
		}
		isStart = true
		stack = &Stack{
			Begin: time.Now(),
		}
		AttachStack(stack)
	}

	var cur *StackEntry
	var oldTop *StackEntry

	file, line := funcInfo.FileLine(pc)

	stack.MaxID++
	cur = &StackEntry{
		ID:       stack.MaxID,
		FuncName: fnName,
		File:     file,
		Line:     line,
		StartNs:  time.Now().UnixNano() - stack.Begin.UnixNano(),
	}
	oldTop = stack.Top
	if oldTop == nil {
		stack.Roots = append(stack.Roots, cur)
	} else {
		cur.ParentID = oldTop.ID
		oldTop.Children = append(oldTop.Children, cur)
	}
	stack.Top = cur
	if stack.OnEnter != nil {
		stack.OnEnter(cur, pc, nil)
	}

	if isStart {
		var outputFile string
		var config runtime.XgoField
		for _, field := range args {
			if field.Name == "config" {
				config = field
				break
			}
		}
		if config.Ptr != nil {
			rvalue := reflect.ValueOf(config.Ptr)
			if rvalue.Kind() == reflect.Ptr {
				rvalue = rvalue.Elem()
			}
			if rvalue.IsValid() && rvalue.Kind() == reflect.Struct {
				outputFileField := rvalue.FieldByName("OutputFile")
				if outputFileField.IsValid() {
					file, ok := outputFileField.Interface().(string)
					if ok {
						outputFile = file
					}
				}
			}
		}
		if outputFile == "" {
			DetachStack()
			return nothing
		}
		stack.OutputFile = outputFile
	}

	// fmt.Fprintf(os.Stderr, "%sargs: %s\n", prefix, string(argsJSON))
	marshalArgs := args
	if recv.Ptr != nil {
		marshalArgs = make([]runtime.XgoField, len(args)+1)
		marshalArgs[0] = recv
		copy(marshalArgs[1:], args)
	}
	cur.Args = json.RawMessage(marshalNoError(StructValue(marshalArgs)))
	stack.Depth++

	return func() {
		cur.EndNs = time.Now().UnixNano() - stack.Begin.UnixNano()

		cur.Results = json.RawMessage(marshalNoError(StructValue(results)))

		stack.Top = oldTop
		stack.Depth--
		// fmt.Fprintf(os.Stderr, "%sreturn %s\n", prefix, fnName)
		if stack.OnExit != nil {
			// TODO: result
			stack.OnExit(cur, pc, nil)
		}
		if oldTop == nil {
			exportedStack := ExportStack(stack)
			exportedStackJSON := marshalNoError(exportedStack)
			err := os.WriteFile(stack.OutputFile, exportedStackJSON, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing stack: %v\n", err)
			}
			// DetachStack()
			// fmt.Fprintf(os.Stderr, "trace end\n")
			DetachStack()
		}
	}
}

func marshalNoError(v interface{}) (result []byte) {
	var err error
	defer func() {
		var stackTrace []byte
		if e := recover(); e != nil {
			stackTrace = debug.Stack()
			if pe, ok := e.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
		var qstackTrace string
		if len(stackTrace) > 0 {
			qstackTrace = fmt.Sprintf(", %q: %q", "stackTrace", string(stackTrace))
		}
		if err != nil {
			result = []byte(fmt.Sprintf("{%q: %q%s}", "error", err.Error(), qstackTrace))
		}
	}()
	result, err = json.Marshal(v)
	return
}

type StructValue []runtime.XgoField

func (c StructValue) MarshalJSON() ([]byte, error) {
	fields := make([]string, len(c))
	for i, v := range c {
		val, err := json.Marshal(v.Ptr)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal field %s: %w", v.Name, err)
		}
		fields[i] = fmt.Sprintf("%q: %s", v.Name, val)
	}
	return []byte(fmt.Sprintf("{%s}", strings.Join(fields, ","))), nil
}
