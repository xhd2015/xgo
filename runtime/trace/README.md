

# Example RPC Interceptor

Data flow:
```
Client add metadata
Server retrieves metadata
```

```go
package trace

import (
	"context"
	"os"
	"path/filepath"

	"go-micro.dev/v4/server"
    "go-micro.dev/v4/metadata"

	"github.com/xhd2015/xgo/runtime/trace"
)

const X_XGO_TRACE_FILE = "X-Xgo-Trace-File"

func NewTraceInterceptor() server.HandlerWrapper {
	return func(hf server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
            // traceFile example: /tmp/some_trace.json
            traceFile, _ := metadata.Get(ctx, X_XGO_TRACE_FILE)
			if traceFile == "" {
				return hf(ctx, req, rsp)
			}

			_, err = trace.Trace(trace.Config{
				OutputFile: traceFile,
			}, req.Body(), func() (interface{}, error) {
				err := hf(ctx, req, rsp)
				return rsp, err
			})
			return err
		}
	}
}
```