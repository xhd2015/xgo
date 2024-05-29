package test_explorer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/netutil"
	"github.com/xhd2015/xgo/support/session"
)

func setupPollHandler(server *http.ServeMux, prefix string, projectDir string, getTestConfig func() (*TestConfig, error), runner func(ctx *RunContext) error) {
	sessionManager := session.NewSessionManager()

	server.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *DebugRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			if req == nil || req.Item == nil || req.Item.File == "" {
				return nil, netutil.ParamErrorf("requires file")
			}
			if req.Item.Name == "" {
				return nil, netutil.ParamErrorf("requires name")
			}

			file := req.Item.File
			isFile, err := fileutil.IsFile(file)
			if err != nil {
				return nil, err
			}
			if !isFile {
				return nil, fmt.Errorf("cannot debug multiple tests")
			}
			absDir, err := filepath.Abs(projectDir)
			if err != nil {
				return nil, err
			}

			parsedFlags, parsedArgs, err := getTestFlags(absDir, file, req.Item.Name)
			if err != nil {
				return nil, err
			}

			relPath, err := filepath.Rel(absDir, file)
			if err != nil {
				return nil, err
			}

			config, err := getTestConfig()
			if err != nil {
				return nil, err
			}

			id, sess, err := sessionManager.Start()
			if err != nil {
				return nil, err
			}

			startRun(sess, projectDir, absDir, file, relPath, req.Item.Name, config, parsedFlags, parsedArgs, runner)
			return &DebugResponse{ID: id}, nil
		})
	})

	server.HandleFunc(prefix+"/pollStatus", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *DebugPollRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			if req.ID == "" {
				return nil, netutil.ParamErrorf("requires id")
			}
			session, err := sessionManager.Get(req.ID)
			if err != nil {
				return nil, err
			}

			events, err := session.PollEvents()
			if err != nil {
				return nil, err
			}
			return &DebugPollResponse{
				Events: convTestingEvents(events),
			}, nil
		})
	})
	server.HandleFunc(prefix+"/destroy", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *DebugDestroyRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			if req.ID == "" {
				return nil, netutil.ParamErrorf("requires id")
			}
			err = sessionManager.Destroy(req.ID)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	})
}

func startRun(sess session.Session, projectDir string, absDir string, file string, relPath string, name string, config *TestConfig, testFlags []string, testArgs []string, runner func(ctx *RunContext) error) {
	pr, pw := io.Pipe()
	// go func() { xxx }
	// - build with gcflags="all=-N -l"
	// - start dlv
	// - output prompt
	go func() {
		defer sess.SendEvents(&TestingItemEvent{
			Event: Event_TestEnd,
		})

		ctx := &RunContext{
			ProjectDir:    projectDir,
			AbsProjectDir: absDir,
			File:          file,
			RelPath:       relPath,
			Name:          name,
			Stdout:        io.MultiWriter(os.Stdout, pw),
			Stderr:        io.MultiWriter(os.Stderr, pw),

			GoCmd:      config.GetGoCmd(),
			BuildFlags: append(config.Flags, testFlags...),
			Env:        config.CmdEnv(),
			Args:       testArgs,
		}
		err := runner(ctx)
		if err != nil {
			sess.SendEvents(&TestingItemEvent{
				Event: Event_Output,
				Msg:   "err: " + err.Error(),
			})
		}
	}()

	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			data := scanner.Bytes()
			sess.SendEvents(&TestingItemEvent{
				Event: Event_Output,
				Msg:   string(data),
			})
		}
	}()
}

type RunContext struct {
	ProjectDir    string
	AbsProjectDir string
	File          string
	RelPath       string
	Name          string
	Stdout        io.Writer
	Stderr        io.Writer

	GoCmd      string
	BuildFlags []string

	Env  []string
	Args []string
}
