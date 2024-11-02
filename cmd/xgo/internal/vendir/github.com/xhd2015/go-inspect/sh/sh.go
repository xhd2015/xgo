package sh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-inspect/sh/process"
)

func RunBash(cmdList []string, verbose bool) error {
	_, _, err := RunBashWithOpts(cmdList, RunBashOptions{
		Verbose: verbose,
	})
	return err
}

type RunBashOptions struct {
	Verbose    bool
	NeedStdErr bool
	NeedStdOut bool

	StdoutNoTrim bool

	Args []string

	ErrExcludeCmd bool
	Timeout       time.Duration

	// if StdoutToJSON != nil, the value is parsed into this struct
	StdoutToJSON  interface{}
	StdoutToBytes *[]byte
	FilterCmd     func(cmd *exec.Cmd)
}

func RunBashWithOpts(cmdList []string, opts RunBashOptions) (stdout string, stderr string, err error) {
	cmdExpr := bashCommandExpr(cmdList)
	if opts.Verbose {
		log.Printf("%s", cmdExpr)
	}
	list := make([]string, 2+len(opts.Args))
	list[0] = "-c"
	list[1] = cmdExpr
	for i, arg := range opts.Args {
		list[i+2] = arg
	}

	// bash -c cmdExpr args...
	cmd := exec.Command("bash", list...)
	stdoutBuf := bytes.NewBuffer(nil)
	stderrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	if opts.FilterCmd != nil {
		opts.FilterCmd(cmd)
	}
	process.SetSysProcAttribute(cmd)

	getCmdDetail := func() string {
		if opts.ErrExcludeCmd {
			return ""
		}
		return fmt.Sprintf("cmd %s ", cmdExpr)

	}

	run := func() (stdout string, stderr string, err error) {
		err = cmd.Run()
		if err != nil {
			cmdDetail := getCmdDetail()
			err = fmt.Errorf("running cmd error: %s%v stdout:%s stderr:%s", cmdDetail, err, stdoutBuf.String(), stderrBuf.String())
			return
		}
		if opts.NeedStdOut {
			stdout = stdoutBuf.String()
			if !opts.StdoutNoTrim {
				stdout = strings.TrimSuffix(stdout, "\n")
			}
		}
		if opts.NeedStdErr {
			stderr = stderrBuf.String()
		}
		if opts.StdoutToBytes != nil {
			*opts.StdoutToBytes = stdoutBuf.Bytes()
		} else if opts.StdoutToJSON != nil {
			err = json.Unmarshal(stdoutBuf.Bytes(), opts.StdoutToJSON)
			if err != nil {
				err = fmt.Errorf("parse command output to %T error:%v", opts.StdoutToJSON, err)
			}
		}
		return
	}

	if opts.Timeout > 0 {
		timeoutCh := time.After(opts.Timeout)
		done := make(chan struct{})
		var subStdout string
		var subStderr string
		var subErr error
		go func() {
			defer func() {
				if e := recover(); e != nil {
					if e := e.(error); e != nil {
						subErr = e
					} else {
						subErr = fmt.Errorf("panic %v", e)
					}
				}
				close(done)
			}()
			subStdout, subStderr, subErr = run()
		}()
		select {
		case <-timeoutCh:
			err = fmt.Errorf("cmd timeout after %v: %s", opts.Timeout, getCmdDetail())
			if cmd.Process != nil {
				// ensure the process is killed
				cmd.Process.Kill()
			}
		case <-done:
			stdout, stderr, err = subStdout, subStderr, subErr
		}
	} else {
		stdout, stderr, err = run()
	}

	return
}

func JoinArgs(args []string) string {
	eArgs := make([]string, 0, len(args))
	for _, arg := range args {
		eArgs = append(eArgs, Quote(arg))
	}
	return strings.Join(eArgs, " ")
}

func Quotes(args ...string) string {
	eArgs := make([]string, 0, len(args))
	for _, arg := range args {
		eArgs = append(eArgs, Quote(arg))
	}
	return strings.Join(eArgs, " ")
}
func Quote(s string) string {
	if s == "" {
		return "''"
	}
	if strings.ContainsAny(s, "\t \n;<>\\${}()&!*") { // special args
		s = strings.ReplaceAll(s, "'", "'\\''")
		return "'" + s + "'"
	}
	return s
}

func bashCommandExpr(cmd []string) string {
	var b strings.Builder
	for i, c := range cmd {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		b.WriteString(c)
		if i >= len(cmd)-1 {
			// last no \n
			continue
		}
		if strings.HasSuffix(c, "\n") || strings.HasSuffix(c, "&&") || strings.HasSuffix(c, ";") || strings.HasSuffix(c, "||") {
			continue
		}
		b.WriteString("\n")
	}
	return strings.Join(cmd, "\n")
}
