package test_explorer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/netutil"
	"github.com/xhd2015/xgo/support/session"
	"github.com/xhd2015/xgo/support/strutil"
)

type DebugRequest struct {
	Item *TestingItem `json:"item"`
}
type DebugResponse struct {
	ID string `json:"id"`
}

type DebugPollRequest struct {
	ID string `json:"id"`
}

type DebugPollResponse struct {
	Events []*TestingItemEvent `json:"events"`
}
type DebugDestroyRequest struct {
	ID string `json:"id"`
}

func setupDebugHandler(server *http.ServeMux, projectDir string, getTestConfig func() (*TestConfig, error)) {
	sessionManager := session.NewSessionManager()

	server.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
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

			id, session, err := sessionManager.Start()
			if err != nil {
				return nil, err
			}

			pr, pw := io.Pipe()

			// go func() { xxx }
			// - build with gcflags="all=-N -l"
			// - start dlv
			// - output prompt
			go func() {
				defer session.SendEvents(&TestingItemEvent{
					Event: Event_TestEnd,
				})
				debug := func(projectDir string, file string, stdout io.Writer, stderr io.Writer) error {
					goCmd := config.GetGoCmd()
					tmpDir, err := os.MkdirTemp("", "go-test-debug")
					if err != nil {
						return err
					}
					defer os.RemoveAll(tmpDir)
					binName := "debug.bin"
					baseName := filepath.Base(file)
					if baseName != "" {
						binName = baseName + "-" + binName
					}

					// TODO: find a way to automatically set breakpoint
					// dlvInitFile := filepath.Join(tmpDir, "dlv-init.txt")
					// err = ioutil.WriteFile(dlvInitFile, []byte(fmt.Sprintf("break %s:%d\n", file, req.Item.Line)), 0755)
					// if err != nil {
					// 	return err
					// }
					relPathDir := filepath.Dir(relPath)
					tmpBin := filepath.Join(tmpDir, binName)

					flags := []string{"test", "-c", "-o", tmpBin, "-gcflags=all=-N -l"}
					flags = append(flags, config.Flags...)
					flags = append(flags, parsedFlags...)
					flags = append(flags, "./"+relPathDir)
					err = cmd.Dir(projectDir).Debug().Stderr(stderr).Stdout(stdout).Run(goCmd, flags...)
					if err != nil {
						return err
					}
					err = netutil.ServePort(2345, true, 500*time.Millisecond, func(port int) {
						// user need to set breakpoint explicitly
						fmt.Fprintf(stderr, "dlv listen on localhost:%d\n", port)
						fmt.Fprintf(stderr, "Debug with IDEs:\n")
						fmt.Fprintf(stderr, "  > VSCode: add the following config to .vscode/launch.json configurations:")
						fmt.Fprintf(stderr, "\n%s\n", strutil.IndentLines(formatVscodeConfig(port), "    "))
						fmt.Fprintf(stderr, "  > GoLand: click Add Configuration > Go Remote > localhost:%d\n", port)
						fmt.Fprintf(stderr, "  > Terminal: dlv connect localhost:%d\n", port)
					}, func(port int) error {
						// dlv exec --api-version=2 --listen=localhost:2345 --accept-multiclient --headless ./debug.bin
						return cmd.Dir(filepath.Dir(file)).Debug().Stderr(stderr).Stdout(stdout).Run("dlv",
							append([]string{
								"exec",
								"--api-version=2",
								"--check-go-version=false",
								// NOTE: --init is ignored if --headless
								// "--init", dlvInitFile,
								"--headless",
								// "--allow-non-terminal-interactive=true",
								fmt.Sprintf("--listen=localhost:%d", port),
								tmpBin, "--", "-test.v", "-test.run", fmt.Sprintf("^%s$", req.Item.Name),
							}, parsedArgs...)...,
						)
					})
					if err != nil {
						return err
					}
					return nil
				}
				err := debug(projectDir, file, io.MultiWriter(os.Stdout, pw), io.MultiWriter(os.Stderr, pw))
				if err != nil {
					session.SendEvents(&TestingItemEvent{
						Event: Event_Output,
						Msg:   "err: " + err.Error(),
					})
				}
			}()

			go func() {
				scanner := bufio.NewScanner(pr)
				for scanner.Scan() {
					data := scanner.Bytes()
					session.SendEvents(&TestingItemEvent{
						Event: Event_Output,
						Msg:   string(data),
					})
				}
			}()
			return &DebugResponse{ID: id}, nil
		})
	})

	server.HandleFunc("/debug/pollStatus", func(w http.ResponseWriter, r *http.Request) {
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
	server.HandleFunc("/debug/destroy", func(w http.ResponseWriter, r *http.Request) {
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

func formatVscodeConfig(port int) string {
	return fmt.Sprintf(`{
	"configurations": [
		{
			"name": "Debug dlv localhost:%d",
			"type": "go",
			"request": "attach",
			"mode": "remote",
			"port": %d,
			"host": "127.0.0.1"
		}
	}
}`, port, port)
}
