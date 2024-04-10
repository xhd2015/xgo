package func_list

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/functab"
)

var _ http.Request
var _ net.Addr
var _ time.Time
var _ exec.Cmd
var _ = ioutil.ReadAll
var _ = ioutil.ReadFile
var _ = ioutil.ReadDir
var _ = io.ReadAll
var _ json.Encoder

// go run ./cmd/xgo test --project-dir runtime -run TestListStdlib -v ./test/func_list
func TestListStdlib(t *testing.T) {
	funcs := functab.GetFuncs()

	stdPkgs := map[string]bool{
		// os
		"os.Getenv":    true,
		"os.Getwd":     true,
		"os.OpenFile":  true,
		"os.ReadFile":  true,
		"os.WriteFile": true,

		// io
		"io.ReadAll": true,

		// io/ioutl
		"io/ioutil.ReadAll":  true,
		"io/ioutil.ReadFile": true,
		"io/ioutil.ReadDir":  true,

		// time
		"time.Now":         true,
		"time.Sleep":       true,
		"time.NewTicker":   true,
		"time.Time.Format": true,

		// exec
		"os/exec.Command":    true,
		"os/exec.Cmd.Run":    true,
		"os/exec.Cmd.Output": true,
		"os/exec.Cmd.Start":  true,

		// http client
		"net/http.Get":       true,
		"net/http.Head":      true,
		"net/http.Post":      true,
		"net/http.Client.Do": true,

		// http server
		"net/http.Serve":        true,
		"net/http.Server.Close": true,
		"net/http.Handle":       true,

		// net
		"net.Dial":        true,
		"net.DialIP":      true,
		"net.DialTCP":     true,
		"net.DialUDP":     true,
		"net.DialUnix":    true,
		"net.DialTimeout": true,

		//json
		"encoding/json.newTypeEncoder": true,
	}
	found, missing := getMissing(funcs, stdPkgs, false)
	if len(missing) > 0 {
		t.Fatalf("expect func list contains: %v, actual %v", missing, found)
	}
}
