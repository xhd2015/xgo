package trace_runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/xhd2015/xgo/runtime/trace/constants"
)

func init() {
	Runtime_XgoSetTrap(trap)
}

func trap(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool) {
	// skip 2: <user func> -> runtime.XgoTrap -> trap
	const SKIP = 2

	var pcs [1]uintptr
	runtime.Callers(SKIP+1, pcs[:])
	pc := pcs[0]
	funcInfo := runtime.FuncForPC(pc)

	fnName := funcInfo.Name()

	var mock func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{})

	var isStart bool

	stack := GetStack()
	if stack != nil {
		if stack.inspecting != nil {
			stack.inspecting(pc, recvName, recvPtr, argNames, args, resultNames, results)
			return nil, true
		}
		wantPtr, mockFn := stack.getLastMock(funcInfo.Entry())
		if mockFn != nil && (wantPtr == nil || (recvPtr != nil && sameReceiver(recvPtr, wantPtr))) {
			mock = mockFn
		}
		if !stack.hasStartedTracing {
			if fnName == constants.START_XGO_TRACE {
				stack.hasStartedTracing = true
				isStart = true
			}
		}
		if !stack.hasStartedTracing {
			// without tracing, mock becomes simpler
			if mock != nil {
				mock(recvName, recvPtr, argNames, args, resultNames, results)
				return nil, true
			}
			return nil, false
		}
	} else {
		if fnName != constants.START_XGO_TRACE {
			return nil, false
		}
		isStart = true
		stack = &Stack{
			Begin:             time.Now(),
			hasStartedTracing: true,
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

	if isStart {
		var outputFile string
		var config interface{}
		for i, arg := range args {
			if argNames[i] == "config" {
				config = arg
				break
			}
		}
		if config != nil {
			rvalue := reflect.ValueOf(config)
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
			return nil, false
		}
		stack.OutputFile = outputFile
	}

	// fmt.Fprintf(os.Stderr, "%sargs: %s\n", prefix, string(argsJSON))
	argNamesNoCtx, argsNoCtx := tryRemoveFirstCtx(argNames, args)
	marshalNames := argNamesNoCtx
	marshalArgs := argsNoCtx
	if recvPtr != nil {
		marshalNames = make([]string, 1+len(argNamesNoCtx))
		marshalArgs = make([]interface{}, 1+len(argsNoCtx))
		marshalNames[0] = recvName
		marshalArgs[0] = recvPtr
		copy(marshalNames[1:], argNamesNoCtx)
		copy(marshalArgs[1:], argsNoCtx)
	}
	cur.Args = json.RawMessage(marshalNoError(newStructValue(marshalNames, marshalArgs)))
	stack.Depth++

	hitMock := mock != nil
	post := func() {
		cur.EndNs = time.Now().UnixNano() - stack.Begin.UnixNano()
		cur.HitMock = hitMock
		var hasPanic bool
		if pe := Runtime_XgoPeekPanic(); pe != nil {
			hasPanic = true
			cur.Panic = true
			cur.Error = fmt.Sprint(pe)
		}

		resultNamesNoErr, resultsNoErr, resErr := trySplitLastError(resultNames, results)
		cur.Results = json.RawMessage(marshalNoError(newStructValue(resultNamesNoErr, resultsNoErr)))
		if !hasPanic && resErr != nil {
			cur.Error = resErr.Error()
		}

		stack.Top = oldTop
		stack.Depth--
		if isStart {
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
	if hitMock {
		defer post()
		mock(recvName, recvPtr, argNames, args, resultNames, results)
		return nil, true
	}
	return post, false
}
