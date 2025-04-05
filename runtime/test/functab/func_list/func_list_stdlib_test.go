package func_list

import (
	_ "encoding/json"
	_ "io"
	_ "io/ioutil"
	_ "net"
	_ "net/http"
	_ "os/exec"
	"testing"
	_ "time"

	"github.com/xhd2015/xgo/runtime/functab"
)

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

		// json
		// "encoding/json.newTypeEncoder": true, // not required since xgo v1.1.0
	}
	// debug
	// stdPkgs = map[string]bool{
	// 	"net/http.Get": true,
	// }
	found, missing := getMissing(funcs, stdPkgs, false)
	if len(missing) > 0 {
		t.Fatalf("expect func list contains: %v, actual %v", missing, found)
	}
}
