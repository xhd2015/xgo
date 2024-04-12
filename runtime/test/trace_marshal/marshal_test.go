package trace_marshal

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestMarshalAnyJSON(t *testing.T) {
	var nilChan chan int
	tests := []struct {
		v    interface{}
		want string
		err  string
	}{
		{
			v:    nil,
			want: "null",
		},
		{
			v:    struct{}{},
			want: "{}",
		},
		{
			v:    func() {},
			want: "{}",
		},
		{
			v:    make(chan int),
			want: "{}",
		},
		{
			v:    nilChan,
			want: "{}",
		},
		{
			v:    struct{ A int }{A: 123},
			want: `{"A":123}`,
		},
		{
			v:    getObject(),
			want: `{"_r0":{}}`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			got, err := trace.MarshalAnyJSON(tt.v)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if (errMsg != "" && tt.err == "") || !strings.Contains(errMsg, tt.err) {
				t.Fatalf("expect err msg: %s, actual: %s", tt.err, errMsg)
			}
			if tt.want != string(got) {
				t.Fatalf("expect result: %s, actual: %s", tt.want, got)
			}
		})
	}
}

type cyclic struct {
	Self *cyclic
	Name string
}

func TestMarshalCyclicJSON(t *testing.T) {
	c := &cyclic{
		Name: "cyclic",
	}
	c.Self = c

	res, err := trace.MarshalAnyJSON(c)
	if err != nil {
		t.Fatal(err)
	}
	resStr := string(res)
	expect := `{"Self":null,"Name":"cyclic"}`
	if resStr != expect {
		t.Fatalf("expect res to be %q, actual: %q", expect, resStr)
	}
}

func exampleReturnFunc() context.CancelFunc {
	_, f := context.WithTimeout(context.TODO(), 10*time.Millisecond)
	return f
}

func getObject() core.Object {
	var recordedResult core.Object
	fnInfo := functab.InfoFunc(exampleReturnFunc)
	trap.WithInterceptor(&trap.Interceptor{Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
		if fnInfo != f {
			return nil, nil
		}
		recordedResult = result
		return nil, nil
	}}, func() {
		exampleReturnFunc()
	})
	return recordedResult
}
