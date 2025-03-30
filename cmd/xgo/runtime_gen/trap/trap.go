package trap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	xgo_runtime "github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/trace/constants"
	"github.com/xhd2015/xgo/runtime/trap/flags"
	"github.com/xhd2015/xgo/runtime/trap/stack_model"
)

func trap(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool) {
	begin := time.Now()
	// skip 2: <user func> -> runtime.XgoTrap -> trap
	const SKIP = 2

	var pcs [1]uintptr
	runtime.Callers(SKIP+1, pcs[:])
	pc := pcs[0]
	funcInfo := runtime.FuncForPC(pc)

	fnName := funcInfo.Name()
	fnPC := funcInfo.Entry()

	var mock func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) bool

	var isStart bool
	var isTesting bool
	var testName string

	var postRecorder func()

	stack := GetStack()
	if stack != nil {
		if stack.inspecting != nil {
			stack.inspecting(pc, recvName, recvPtr, argNames, args, resultNames, results)
			return nil, true
		}
		wantPtr, mockFn := stack.getLastMock(fnPC)
		if mockFn != nil && (wantPtr == nil || (recvPtr != nil && sameReceiver(recvPtr, wantPtr))) {
			mock = mockFn
		}
		if !stack.hasStartedTracing {
			if fnName == constants.START_XGO_TRACE {
				stack.hasStartedTracing = true
				isStart = true
			}
		}
		var postRecorders []func()
		recordHandlers := stack.getRecordHandlers(fnPC)
		for _, h := range recordHandlers {
			if h.wantRecvPtr != nil && (recvPtr == nil || !sameReceiver(recvPtr, h.wantRecvPtr)) {
				continue
			}
			var data interface{}
			if h.pre != nil {
				data, _ = h.pre(recvName, recvPtr, argNames, args, resultNames, results)
			}
			if h.post != nil {
				postRecorders = append(postRecorders, func() {
					h.post(recvName, recvPtr, argNames, args, resultNames, results, data)
				})
			}
		}
		if len(postRecorders) > 0 {
			if len(postRecorders) == 1 {
				postRecorder = postRecorders[0]
			} else {
				postRecorder = func() {
					// reversed
					n := len(postRecorders)
					for i := n - 1; i >= 0; i-- {
						postRecorders[i]()
					}
				}
			}
		}
		if !stack.hasStartedTracing {
			// without tracing, mock becomes simpler
			if mock != nil {
				ok := mock(recvName, recvPtr, argNames, args, resultNames, results)
				// ok=true indicates not call old function
				return postRecorder, ok
			}
			return postRecorder, false
		}
	} else {
		if fnName != constants.START_XGO_TRACE {
			if !flags.COLLECT_TEST_TRACE {
				return nil, false
			}
			if !(recvPtr == nil && len(args) == 1 && len(results) == 0) {
				return nil, false
			}
			t, ok := args[0].(**testing.T)
			if !ok {
				return nil, false
			}
			// detect if we are called from TestX(t *testing.T)
			var pcs [1]uintptr
			runtime.Callers(SKIP+2, pcs[:])
			pc := pcs[0]
			funcInfo := runtime.FuncForPC(pc)

			fnName := funcInfo.Name()
			if fnName != "testing.tRunner" {
				return nil, false
			}
			isTesting = true
			testName = (*t).Name()
		}
		isStart = true
		stack = &Stack{
			Begin:             begin,
			hasStartedTracing: true,
		}
		AttachStack(stack)
	}

	file, line := funcInfo.FileLine(pc)
	cur := stack.newStackEntry(begin, fnName)
	oldTop := stack.push(cur)
	cur.File = file
	cur.Line = line

	if isStart && !isTesting {
		var onFinish func(stack stack_model.IStack)
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
				onFinishField := rvalue.FieldByName("OnFinish")
				if onFinishField.IsValid() {
					f, ok := onFinishField.Interface().(func(stack stack_model.IStack))
					if ok {
						onFinish = f
					}
				}
			}
		}
		if outputFile == "" && onFinish == nil {
			DetachStack()
			return postRecorder, false
		}
		stack.OutputFile = outputFile
		stack.onFinish = onFinish
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

	var hitMock bool
	post := func() {
		if postRecorder != nil {
			postRecorder()
		}
		cur.EndNs = time.Now().UnixNano() - stack.Begin.UnixNano()
		cur.HitMock = hitMock
		var hasPanic bool
		if pe := xgo_runtime.XgoPeekPanic(); pe != nil {
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
			exportedStack := ExportStack(stack, 0)
			exportedStackJSON := marshalNoError(exportedStack)
			if isTesting {
				outputFile := filepath.Join(flags.COLLECT_TEST_TRACE_DIR, testName+".json")
				os.MkdirAll(filepath.Dir(outputFile), 0755)
				err := os.WriteFile(outputFile, exportedStackJSON, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error writing stack: %v\n", err)
				}
			} else {
				if stack.onFinish != nil {
					stack.onFinish(&stackData{
						data: exportedStack,
						json: exportedStackJSON,
					})
				}
				if stack.OutputFile != "" {
					err := os.WriteFile(stack.OutputFile, exportedStackJSON, 0644)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error writing stack: %v\n", err)
					}
				}
			}
			// DetachStack()
			// fmt.Fprintf(os.Stderr, "trace end\n")
			DetachStack()
		}
	}
	if mock != nil {
		defer post()
		hitMock = mock(recvName, recvPtr, argNames, args, resultNames, results)
		return nil, hitMock
	}
	return post, false
}

type stackData struct {
	data *stack_model.Stack
	json []byte
}

func (c *stackData) Data() *stack_model.Stack {
	return c.data
}

func (c *stackData) JSON() ([]byte, error) {
	return c.json, nil
}

// push returns the old top
func (c *Stack) push(cur *StackEntry) *StackEntry {
	c.MaxID++
	oldTop := c.Top
	if oldTop == nil {
		c.Roots = append(c.Roots, cur)
	} else {
		cur.ParentID = oldTop.ID
		oldTop.Children = append(oldTop.Children, cur)
	}
	c.Top = cur
	return oldTop
}

func (c *Stack) newStackEntry(begin time.Time, fnName string) *StackEntry {
	c.MaxID++
	cur := &StackEntry{
		ID:       c.MaxID,
		FuncName: fnName,
		StartNs:  begin.UnixNano() - c.Begin.UnixNano(),
	}
	return cur
}
