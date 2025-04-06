package trap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core/info"
	"github.com/xhd2015/xgo/runtime/internal/constants"
	"github.com/xhd2015/xgo/runtime/internal/flags"
	xgo_runtime "github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/internal/stack"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

// skip 2: <user func> -> runtime.XgoTrap -> trap
const SKIP = 2

// trap is the function called upon every go function call,
// it implements mock, recording and tracing functionality
// the trap discarded the interceptor design in xgo v1.0,
// and uses a simpler and more efficient design:
//   - mapping by pc and variable pointer
//
// this avoids the infinite trap problem
func trap(infoPtr unsafe.Pointer, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool) {
	funcInfo := (*info.Func)(infoPtr)
	recvName := funcInfo.RecvName
	argNames := funcInfo.ArgNames
	resultNames := funcInfo.ResNames

	begin := xgo_runtime.XgoRealTimeNow()

	var pcs [1]uintptr
	runtime.Callers(SKIP+1, pcs[:])
	pc := pcs[0]
	runtimeFuncInfo := runtime.FuncForPC(pc)
	fnPC := runtimeFuncInfo.Entry()

	pkg := funcInfo.Pkg
	name := funcInfo.IdentityName

	var mock func(fnInfo *info.Func, recvPtr interface{}, args []interface{}, results []interface{}) bool

	var isStartTracing bool
	var isTesting bool
	var testName string

	var postRecorder func()

	stk := stack.Get()
	if stk == stack.NilGStack {
		return nil, false
	}
	stackData := getStackDataOf(stk)
	if stackData != nil {
		if stackData.inspecting != nil {
			stackData.inspecting(pc, funcInfo, recvPtr, args, results)
			return nil, true
		}

		stackIsTrapping := stackData.handlingTrapping
		if !stackIsTrapping {
			stackData.handlingTrapping = true
			defer func() {
				stackData.handlingTrapping = false
			}()
		}

		wantPtr, mockFn := stackData.getLastMock(fnPC)
		if mockFn != nil && (wantPtr == nil || (recvPtr != nil && sameReceiver(recvPtr, wantPtr))) {
			mock = mockFn
		}
		if !stackIsTrapping && !stackData.hasStartedTracing {
			if pkg == constants.TRACE_PKG && name == constants.TRACE_FUNC {
				stackData.hasStartedTracing = true
				isStartTracing = true
			}
		}
		var postRecordersAndInterceptors []func()
		recordHandlers := stackData.getRecordHandlers(fnPC)
		for _, h := range recordHandlers {
			if h.wantRecvPtr != nil && (recvPtr == nil || !sameReceiver(recvPtr, h.wantRecvPtr)) {
				continue
			}
			var data interface{}
			if h.pre != nil {
				data, _ = h.pre(funcInfo, recvPtr, args, results)
			}
			if h.post != nil {
				postRecordersAndInterceptors = append(postRecordersAndInterceptors, func() {
					h.post(funcInfo, recvPtr, args, results, data)
				})
			}
		}

		// when stack is trapping, we cannot not
		// call into interceptors which are not
		// targeting specific functions, can
		// cause infinite loop.
		// mock and recorders do not have such
		// problem because they explicitly have
		// targeted function
		if !stackIsTrapping {
			interceptors := stackData.interceptors
			for _, h := range interceptors {
				var data interface{}
				if h.pre != nil {
					// TODO: handle abort
					data, _ = h.pre(funcInfo, recvPtr, args, results)
				}

				if h.post != nil {
					postRecordersAndInterceptors = append(postRecordersAndInterceptors, func() {
						h.post(funcInfo, recvPtr, args, results, data)
					})
				}
			}
		}
		if len(postRecordersAndInterceptors) > 0 {
			if len(postRecordersAndInterceptors) == 1 {
				postRecorder = postRecordersAndInterceptors[0]
			} else {
				postRecorder = func() {
					// reversed
					n := len(postRecordersAndInterceptors)
					for i := n - 1; i >= 0; i-- {
						postRecordersAndInterceptors[i]()
					}
				}
			}
		}

		var callPosRecorder func()
		if postRecorder != nil {
			callPosRecorder = func() {
				if !stackData.handlingTrapping {
					stackData.handlingTrapping = true
					defer func() {
						stackData.handlingTrapping = false
					}()
				}
				postRecorder()
			}
		}
		if stackIsTrapping {
			// when stack is trapping, only allow pc-related
			// mock and recorders to run
			if mock != nil {
				ok := mock(funcInfo, recvPtr, args, results)
				// ok=true indicates not call old function
				return callPosRecorder, ok
			}
			return callPosRecorder, false
		}
		if !stackData.hasStartedTracing {
			// without tracing, mock becomes simpler
			if mock != nil {
				ok := mock(funcInfo, recvPtr, args, results)
				// ok=true indicates not call old function
				return callPosRecorder, ok
			}
			return callPosRecorder, false
		}
	} else {
		if stk != nil {
			// this should never happen
			panic("stack is nil while stackData is not nil!")
		}
		if pkg != constants.TRACE_PKG || name != constants.TRACE_FUNC {
			// try detect testing
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

			if funcInfo == nil || funcInfo.Name() != constants.TESTING_RUNNER {
				return nil, false
			}
			isTesting = true
			testName = (*t).Name()
		}
		isStartTracing = true
		stackData = &StackData{
			handlingTrapping:  true,
			hasStartedTracing: true,
		}
		defer func() {
			stackData.handlingTrapping = false
		}()
		stk = &stack.Stack{
			Begin: begin,
			Data: map[interface{}]interface{}{
				dataKey: stackData,
			},
		}
		stack.Attach(stk)
	}

	file, line := runtimeFuncInfo.FileLine(pc)
	cur := stk.NewEntry(begin, name)
	oldTop := stk.Push(cur)
	cur.File = file
	cur.Line = line
	cur.FuncInfo = funcInfo

	if isStartTracing && !isTesting {
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
			stack.Detach()
			return postRecorder, false
		}
		stackData.stackOutputFile = outputFile
		stackData.onFinish = onFinish
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
	cur.Args = json.RawMessage(xgo_runtime.MarshalNoError(newStructValue(marshalNames, marshalArgs)))
	stk.Depth++

	var hitMock bool
	post := func() {
		if !stackData.handlingTrapping {
			stackData.handlingTrapping = true
			defer func() {
				stackData.handlingTrapping = false
			}()
		}
		// NOTE: this defer might be executed on system stack
		// so cannot defer
		if postRecorder != nil {
			postRecorder()
		}
		// on Windows, short stack might resolve to same
		// nanosecond
		// see https://github.com/xhd2015/xgo/issues/307
		// so we add a standalone flag `Finished`
		end := xgo_runtime.XgoRealTimeNow()
		cur.EndNs = end.UnixNano() - stk.Begin.UnixNano()
		cur.Finished = true
		cur.HitMock = hitMock
		var hasPanic bool
		if pe := xgo_runtime.XgoPeekPanic(); pe != nil {
			hasPanic = true
			cur.Panic = true
			cur.Error = fmt.Sprint(pe)
		}

		resultNamesNoErr, resultsNoErr, resErr := trySplitLastError(resultNames, results)
		cur.Results = json.RawMessage(xgo_runtime.MarshalNoError(newStructValue(resultNamesNoErr, resultsNoErr)))
		if !hasPanic && resErr != nil {
			cur.Error = resErr.Error()
		}

		stk.Top = oldTop
		stk.Depth--
		if isStartTracing {
			stk.End = end
			exportedStack := stack.Export(stk, 0)
			exportedStackJSON := xgo_runtime.MarshalNoError(exportedStack)
			if isTesting {
				outputFile := filepath.Join(flags.COLLECT_TEST_TRACE_DIR, testName+".json")
				os.MkdirAll(filepath.Dir(outputFile), 0755)
				err := os.WriteFile(outputFile, exportedStackJSON, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error writing stack: %v\n", err)
				}
			} else {
				if stackData.onFinish != nil {
					stackData.onFinish(&StackDataExportImpl{
						data: exportedStack,
						json: exportedStackJSON,
					})
				}
				if stackData.stackOutputFile != "" {
					err := os.WriteFile(stackData.stackOutputFile, exportedStackJSON, 0644)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error writing stack: %v\n", err)
					}
				}
			}
			// DetachStack()
			// fmt.Fprintf(os.Stderr, "trace end\n")
			stack.Detach()
		}
	}
	if mock != nil {
		defer post()
		hitMock = mock(funcInfo, recvPtr, args, results)
		return nil, hitMock
	}
	return post, false
}

type StackDataExportImpl struct {
	data *stack_model.Stack
	json []byte
}

func (c *StackDataExportImpl) Data() *stack_model.Stack {
	return c.data
}

func (c *StackDataExportImpl) JSON() ([]byte, error) {
	return c.json, nil
}
