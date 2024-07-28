package test_explorer

import (
	"time"
)

type MockStatus string

const (
	MockStatus_NormalResp  MockStatus = "normal_resp"
	MockStatus_NormalError MockStatus = "normal_error"
	MockStatus_MockResp    MockStatus = "mock_resp"
	MockStatus_MockError   MockStatus = "mock_error"
)

type RootRecord struct {
	Start time.Time // the absolute begin time
	Root  *CallRecord
}

type CallRecord struct {
	Pkg        string        `json:"pkg"`
	Func       string        `json:"func"`
	File       string        `json:"file"`
	Line       int           `json:"line"`  // 1-based
	Start      int64         `json:"start"` // relative to request begin, as nanoseconds
	End        int64         `json:"end"`
	Args       interface{}   `json:"args"`
	MockStatus MockStatus    `json:"mockStatus"`
	Error      string        `json:"error"`  // has error, may be empty
	Panic      bool          `json:"panic"`  // has panic
	Result     interface{}   `json:"result"` // keyed by name, if no name, a slice
	Log        interface{}   `json:"log"`    // log set within request
	Children   []*CallRecord `json:"children"`
}

func convertTraceFileToCallRecords() []*CallRecord {
	return nil
}

type traceConverter struct {
}

func (c *traceConverter) convertRoot(root *RootExport) *RootRecord {
	if root == nil {
		return nil
	}
	children := c.convertStacks(root.Children)
	record := &RootRecord{
		Start: root.Begin,
		Root: &CallRecord{
			Children: children,
		},
	}

	// fill root
	if len(children) > 0 {
		firstChild := children[0]
		lastChild := children[len(children)-1]
		record.Root.Start = firstChild.Start
		record.Root.End = lastChild.End

		record.Root.Error = lastChild.Error
		record.Root.Panic = lastChild.Panic
		record.Root.Args = firstChild.Args
		record.Root.Result = lastChild.Result
	}

	return record
}

func (c *traceConverter) convertStacks(stacks []*StackExport) []*CallRecord {
	if stacks == nil {
		return nil
	}
	n := len(stacks)
	convStacks := make([]*CallRecord, n)
	for i := 0; i < n; i++ {
		convStacks[i] = c.convertStack(stacks[i])
	}
	return convStacks
}

func (c *traceConverter) convertStack(stack *StackExport) *CallRecord {
	if stack == nil {
		return nil
	}
	funcInfo := stack.FuncInfo
	if funcInfo == nil {
		funcInfo = &FuncInfoExport{}
	}
	var args interface{} = stack.Args
	var results interface{} = stack.Results
	file := funcInfo.File
	// stack.Args
	return &CallRecord{
		Pkg:  funcInfo.Pkg,
		Func: funcInfo.IdentityName,
		File: file,
		Line: funcInfo.Line,

		Start: stack.Begin,
		End:   stack.End,

		Args: args,

		Error: stack.Error,
		Panic: stack.Panic,

		Result: results,

		Children: c.convertStacks(stack.Children),
	}
}
