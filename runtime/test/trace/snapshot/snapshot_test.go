package snapshot

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/support/assert"
)

// NOTE: do not run with -cover, because
// extra function will be included in trace

// TODO: this test currently fails, but
// it is not so urgent to fix it, so
// I'll keep it here.
// once I got enough time to do, I'll
// try my best.
// see related https://github.com/xhd2015/xgo/issues/281
func TestNoSnapshot(t *testing.T) {
	test(t, noSnapshotExpect, nil)
}

func TestSnapshot(t *testing.T) {
	test(t, snapshotExpect, func(stack *trace.Stack) bool {
		return true
	})
}

func test(t *testing.T, expectTrace string, f func(stack *trace.Stack) bool) {
	var record *trace.Root
	opts := trace.Options()
	if f != nil {
		opts.WithSnapshot(f)
	}
	opts.OnComplete(func(root *trace.Root) {
		record = root
	}).Collect(func() {
		handleL1()
	})
	exportRoot := record.Export(nil)
	// zero
	exportRoot.Begin = time.Time{}
	for _, child := range exportRoot.Children {
		stabilizeTrace(child)
	}

	data, err := trace.MarshalAnyJSON(exportRoot)
	if err != nil {
		t.Error(err)
		return
	}

	var v *trace.RootExport
	err = json.Unmarshal([]byte(expectTrace), &v)
	if err != nil {
		t.Error(err)
		return
	}

	expectBytes, err := json.Marshal(v)
	if err != nil {
		t.Error(err)
		return
	}

	dataStr := string(data)
	expect := string(expectBytes)
	// t.Logf("trace: %s", dataStr)
	if diff := assert.Diff(expect, dataStr); diff != "" {
		t.Errorf("trace not match: %s", diff)
	}
}

func stabilizeTrace(t *trace.StackExport) {
	if t == nil {
		return
	}
	t.Begin = 0
	t.End = 0
	t.FuncInfo.Line = 0
	t.FuncInfo.File = filepath.Base(t.FuncInfo.File)

	for _, child := range t.Children {
		stabilizeTrace(child)
	}
}

const snapshotExpect = `{
    "Begin": "0001-01-01T00:00:00Z",
    "Children": [
        {
            "FuncInfo": {
                "Kind": "func",
                "Pkg": "github.com/xhd2015/xgo/runtime/test/trace/snapshot",
                "IdentityName": "test.func2",
                "Name": "",
                "RecvType": "",
                "RecvPtr": false,
                "Interface": false,
                "Generic": false,
                "Closure": true,
                "Stdlib": false,
                "File": "snapshot_test.go",
                "Line": 0,
                "RecvName": "",
                "ArgNames": null,
                "ResNames": null,
                "FirstArgCtx": false,
                "LastResultErr": false
            },
            "Begin": 0,
            "End": 0,
            "Args": {},
            "Results": {},
            "Snapshot": true,
            "Panic": false,
            "Error": "",
            "Children": [
                {
                    "FuncInfo": {
                        "Kind": "func",
                        "Pkg": "github.com/xhd2015/xgo/runtime/test/trace/snapshot",
                        "IdentityName": "handleL1",
                        "Name": "handleL1",
                        "RecvType": "",
                        "RecvPtr": false,
                        "Interface": false,
                        "Generic": false,
                        "Closure": false,
                        "Stdlib": false,
                        "File": "snapshot.go",
                        "Line": 0,
                        "RecvName": "",
                        "ArgNames": null,
                        "ResNames": [
                            "_r0"
                        ],
                        "FirstArgCtx": false,
                        "LastResultErr": false
                    },
                    "Begin": 0,
                    "End": 0,
                    "Args": {},
                    "Results": {
                        "_r0": {
                            "Last": "handleL1",
                            "Count": 2
                        }
                    },
                    "Snapshot": true,
                    "Panic": false,
                    "Error": "",
                    "Children": [
                        {
                            "FuncInfo": {
                                "Kind": "func",
                                "Pkg": "github.com/xhd2015/xgo/runtime/test/trace/snapshot",
                                "IdentityName": "handleL2",
                                "Name": "handleL2",
                                "RecvType": "",
                                "RecvPtr": false,
                                "Interface": false,
                                "Generic": false,
                                "Closure": false,
                                "Stdlib": false,
                                "File": "snapshot.go",
                                "Line": 0,
                                "RecvName": "",
                                "ArgNames": null,
                                "ResNames": [
                                    "_r0",
                                    "_r1"
                                ],
                                "FirstArgCtx": false,
                                "LastResultErr": false
                            },
                            "Begin": 0,
                            "End": 0,
                            "Args": {},
                            "Results": {
                                "_r0": {
                                    "Last": "handleL3",
                                    "Count": 1
                                },
                                "_r1": {
                                    "Last": "handleL2",
                                    "Count": 1
                                }
                            },
                            "Snapshot": true,
                            "Panic": false,
                            "Error": "",
                            "Children": [
                                {
                                    "FuncInfo": {
                                        "Kind": "func",
                                        "Pkg": "github.com/xhd2015/xgo/runtime/test/trace/snapshot",
                                        "IdentityName": "handleL3",
                                        "Name": "handleL3",
                                        "RecvType": "",
                                        "RecvPtr": false,
                                        "Interface": false,
                                        "Generic": false,
                                        "Closure": false,
                                        "Stdlib": false,
                                        "File": "snapshot.go",
                                        "Line": 0,
                                        "RecvName": "",
                                        "ArgNames": [
                                            "a"
                                        ],
                                        "ResNames": [
                                            "_r0"
                                        ],
                                        "FirstArgCtx": false,
                                        "LastResultErr": false
                                    },
                                    "Begin": 0,
                                    "End": 0,
                                    "Args": {
                                        "a": {
                                            "Last": "handleL2",
                                            "Count": 0
                                        }
                                    },
                                    "Results": {
                                        "_r0": {
                                            "Last": "handleL3",
                                            "Count": 0
                                        }
                                    },
                                    "Snapshot": true,
                                    "Panic": false,
                                    "Error": "",
                                    "Children": null
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}`

// NOTE in this original output without snapshot, handleL3's input.Last = handleL3,while it
// should be handleL2 if snapshot
const noSnapshotExpect = `{
    "Begin": "0001-01-01T00:00:00Z",
    "Children": [
        {
            "FuncInfo": {
                "Kind": "func",
                "Pkg": "github.com/xhd2015/xgo/runtime/test/trace/snapshot",
                "IdentityName": "test.func2",
                "Name": "",
                "RecvType": "",
                "RecvPtr": false,
                "Interface": false,
                "Generic": false,
                "Closure": true,
                "Stdlib": false,
                "File": "snapshot_test.go",
                "Line": 0,
                "RecvName": "",
                "ArgNames": null,
                "ResNames": null,
                "FirstArgCtx": false,
                "LastResultErr": false
            },
            "Begin": 0,
            "End": 0,
            "Args": {},
            "Results": {},
            "Panic": false,
            "Error": "",
            "Children": [
                {
                    "FuncInfo": {
                        "Kind": "func",
                        "Pkg": "github.com/xhd2015/xgo/runtime/test/trace/snapshot",
                        "IdentityName": "handleL1",
                        "Name": "handleL1",
                        "RecvType": "",
                        "RecvPtr": false,
                        "Interface": false,
                        "Generic": false,
                        "Closure": false,
                        "Stdlib": false,
                        "File": "snapshot.go",
                        "Line": 0,
                        "RecvName": "",
                        "ArgNames": null,
                        "ResNames": [
                            "_r0"
                        ],
                        "FirstArgCtx": false,
                        "LastResultErr": false
                    },
                    "Begin": 0,
                    "End": 0,
                    "Args": {},
                    "Results": {
                        "_r0": {
                            "Last": "handleL1",
                            "Count": 2
                        }
                    },
                    "Panic": false,
                    "Error": "",
                    "Children": [
                        {
                            "FuncInfo": {
                                "Kind": "func",
                                "Pkg": "github.com/xhd2015/xgo/runtime/test/trace/snapshot",
                                "IdentityName": "handleL2",
                                "Name": "handleL2",
                                "RecvType": "",
                                "RecvPtr": false,
                                "Interface": false,
                                "Generic": false,
                                "Closure": false,
                                "Stdlib": false,
                                "File": "snapshot.go",
                                "Line": 0,
                                "RecvName": "",
                                "ArgNames": null,
                                "ResNames": [
                                    "_r0",
                                    "_r1"
                                ],
                                "FirstArgCtx": false,
                                "LastResultErr": false
                            },
                            "Begin": 0,
                            "End": 0,
                            "Args": {},
                            "Results": {
                                "_r0": {
                                    "Last": "handleL3",
                                    "Count": 1
                                },
                                "_r1": {
                                    "Last": "handleL1",
                                    "Count": 2
                                }
                            },
                            "Panic": false,
                            "Error": "",
                            "Children": [
                                {
                                    "FuncInfo": {
                                        "Kind": "func",
                                        "Pkg": "github.com/xhd2015/xgo/runtime/test/trace/snapshot",
                                        "IdentityName": "handleL3",
                                        "Name": "handleL3",
                                        "RecvType": "",
                                        "RecvPtr": false,
                                        "Interface": false,
                                        "Generic": false,
                                        "Closure": false,
                                        "Stdlib": false,
                                        "File": "snapshot.go",
                                        "Line": 0,
                                        "RecvName": "",
                                        "ArgNames": [
                                            "a"
                                        ],
                                        "ResNames": [
                                            "_r0"
                                        ],
                                        "FirstArgCtx": false,
                                        "LastResultErr": false
                                    },
                                    "Begin": 0,
                                    "End": 0,
                                    "Args": {
                                        "a": {
                                            "Last": "handleL3",
                                            "Count": 1
                                        }
                                    },
                                    "Results": {
                                        "_r0": {
                                            "Last": "handleL1",
                                            "Count": 2
                                        }
                                    },
                                    "Panic": false,
                                    "Error": "",
                                    "Children": null
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}`
