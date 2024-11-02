package vscode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// removed go:generate bash gen.sh

var initOnce sync.Once
var diffFile string
var cmd *exec.Cmd

var cmdOut *output
var cmdIn *input

// flags
var disableDebugLog = true

var id int64 = 0

// first id is 1
func nextID() string {
	nextID := atomic.AddInt64(&id, 1)
	return strconv.FormatInt(nextID, 10)
}

// ID(string) -> chan string (line channel)
var respMap sync.Map

type input struct {
	lineCh   chan string // input ch
	lastLine string
	lastLF   bool
}

func (c *input) Read(p []byte) (n int, err error) {
	writeLine := func(line string, p []byte, lastLF bool) (leftLine string, nextLastLF bool, n int, done bool) {
		max := len(p)
		maxL := len(line)

		for ; n < max && n < maxL; n++ {
			p[n] = line[n]
		}
		nextLastLF = lastLF
		if maxL >= max {
			leftLine = line[max:]
			done = true
			return
		}
		leftLine = ""
		if lastLF {
			p[n] = '\n'
			n++
			nextLastLF = false
			done = n >= max
		}
		return
	}

	var done bool
	c.lastLine, c.lastLF, n, done = writeLine(c.lastLine, p, c.lastLF)
	if done {
		return
	}
	p = p[n:]
	// c.lastLine must be "", lf must false
	for {
		select {
		case line := <-c.lineCh:
			// fmt.Printf("on line:%s\n", line)
			var m int
			c.lastLine, c.lastLF, m, done = writeLine(line, p, true)
			n += m
			if done {
				return
			}
			p = p[m:]
		case <-time.After(10 * time.Millisecond):
			// if no data detected after 10ms, return
			return
		}
	}
}

type output struct {
	idBuf      []byte
	contentBuf []byte
	respCh     chan respLine
	state      OutputState
}
type OutputState int

var (
	OutputStateID OutputState = 0
	OutputContent OutputState = 1
)

type respLine struct {
	ID      string
	Content string
}

func (c *output) Write(p []byte) (n int, err error) {
	n = len(p)
	for i := 0; i < n; i++ {
		switch c.state {
		case OutputStateID:
			if p[i] != ':' {
				c.idBuf = append(c.idBuf, p[i])
			} else {
				// ignore :
				c.state = OutputContent
			}
		case OutputContent:
			if p[i] != '\n' {
				c.contentBuf = append(c.contentBuf, p[i])
			} else {
				c.respCh <- respLine{ID: string(c.idBuf), Content: string(c.contentBuf)}
				c.idBuf = make([]byte, 0, 10)
				c.contentBuf = make([]byte, 0, 10240) // 10K
				c.state = OutputStateID
			}
		default:
			panic(fmt.Errorf("unrecognized state: %v", c.state))
		}
	}
	return
}

func initProcess() {
	initOnce.Do(func() {
		doInitProcessV1("diff.js", diffJSCode)
	})
}

func doInitProcessV1(file string, code string) {
	var err error
	diffFile, err = InitCode(file, code)
	if err != nil {
		panic(err)
	}
	cmd = exec.Command("node", diffFile)
	cmdIn = &input{
		lineCh: make(chan string, 10),
	}
	cmdOut = &output{
		respCh: make(chan respLine, 10),
	}

	// cmd.Stdin = bytes.NewReader([]byte("{}"))
	cmd.Stdin = cmdIn
	cmd.Stdout = cmdOut
	cmd.Stderr = os.Stderr
	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = append(cmd.Env, "RESPONSE_ID_PREFIX=true")
	cmd.Env = append(cmd.Env, "EXIT_AFTER_PING_TIMEOUT=true")
	if disableDebugLog {
		cmd.Env = append(cmd.Env, "DISABLE_DEBUG_LOG=true")
	}
	// cmd.Stdout = os.Stdout
	err = cmd.Start()
	if err != nil {
		panic(fmt.Errorf("start node err: %v %v", cmd, err))
	}

	// ping
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			cmdIn.lineCh <- `{"ping":true}`
		}
	}()

	// retrieve result
	go func() {
		defer func() {
			close(cmdIn.lineCh)
			close(cmdOut.respCh)
		}()
		for resp := range cmdOut.respCh {
			if val, ok := respMap.Load(resp.ID); ok {
				if ch, ok := val.(chan string); ok {
					ch <- resp.Content
				}
			}
		}
	}()
}

type Request struct {
	OldLines []string `json:"oldLines"`
	NewLines []string `json:"newLines"`
	// Pretty bool `json:"pretty"` // should default to true get intrituive difff
	// ComputeChar
	// pretty?: boolean // default true
	// computeChar?: boolean // default false
}

type Result struct {
	QuitEarly bool          `json:"quitEarly"`
	Changes   []*LineChange `json:"changes"`
}

type internRequest struct {
	ID string `json:"id"` // request id, override automatcally by Diff, do not set.
	*Request
}

type internResult struct {
	Error string `json:"error,omitempty"` // if any inner error
	*Result
}

type LineChange struct {
	OriginalStartLineNumber int `json:"originalStartLineNumber"`
	OriginalEndLineNumber   int `json:"originalEndLineNumber"` // 0: insertion
	ModifiedStartLineNumber int `json:"modifiedStartLineNumber"`
	ModifiedEndLineNumber   int `json:"modifiedEndLineNumber"` // 0: delete
}

// compressed representation
// type LineChangeLite [4]int // 4

// DestryNow terminates subprocess immediately, rather than waiting for 10s.
// If not called it will terminate automatically after 10s.
func DestroyNow() error {
	return cmd.Process.Kill()
}

// There is a bottleneck here:
//   because we only spawn 1 process, all requests and resposne are serial(though they may reordered),
//

// TODO: may spawn as many process as possible
func DiffV1(req *Request) (*Result, error) {
	initProcess()

	ireq := &internRequest{
		ID:      nextID(),
		Request: req,
	}
	// ID to solve concurrent problems.
	reqLine, err := json.Marshal(ireq)
	if err != nil {
		return nil, err
	}
	cmdIn.lineCh <- string(reqLine)

	// make chan and store to respMap
	ch := make(chan string)
	respMap.Store(ireq.ID, ch)
	defer func() {
		respMap.Delete(ireq.ID)
		close(ch)
	}()

	// timeout after 1s
	var outLine string
	select {
	case outLine = <-ch:
		// pass
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("diff result timeout after 5s: requestID=%v", ireq.ID)
	}

	var res internResult
	err = json.Unmarshal([]byte(outLine), &res)
	if err != nil {
		return nil, fmt.Errorf("decoding result err: len=%v %v", len(outLine), err)
	}
	if res.Error != "" {
		return nil, errors.New(res.Error)
	}
	return res.Result, nil
}
