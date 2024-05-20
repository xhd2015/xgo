package netutil

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

type HttpStatusErr interface {
	error
	HttpStatusCode() int
}

type badParamErr struct {
	msg string
}

func (c *badParamErr) HttpStatusCode() int {
	return 400
}

func (c *badParamErr) Error() string {
	return c.msg
}

func ParamErrorf(format string, args ...interface{}) error {
	return &badParamErr{msg: fmt.Sprintf(format, args...)}
}

func HandleJSON(w http.ResponseWriter, r *http.Request, h func(ctx context.Context, r *http.Request) (interface{}, error)) {
	if r.Method == http.MethodOptions {
		return
	}
	var respData interface{}
	var err error
	defer func() {
		if e := recover(); e != nil {
			// print panic stack trace
			stack := debug.Stack()
			log.Printf("panic: %s", stack)
			if pe, ok := e.(error); ok {
				e = pe
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
		var jsonData []byte
		if err == nil {
			jsonData, err = json.Marshal(respData)
		}

		if err != nil {
			log.Printf("err: %v", err)
			code := 500
			if httpErr, ok := err.(HttpStatusErr); ok {
				code = httpErr.HttpStatusCode()
			}
			w.WriteHeader(code)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}()

	respData, err = h(context.Background(), r)
	if err != nil {
		return
	}
}

// allow request from arbitrary host
func SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}
