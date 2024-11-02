package coverage

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/coverage/serve"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/git"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/load/loadcov"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/netutil"
)

func handleServe(cmd string, args []string) error {
	var diffWith string
	var bind string
	var port string = "8000"
	n := len(args)
	var remain []string
	var projectDir string
	var buildArgs []string

	var include []string
	var exclude []string
	var full bool
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			remain = append(remain, args[i+1:]...)
			break
		}
		if arg == "--bind" {
			if i+1 > n {
				return fmt.Errorf("%s requires arg", arg)
			}
			bind = args[i+1]
			i++
			continue
		}
		if arg == "--port" {
			if i+1 > n {
				return fmt.Errorf("%s requires arg", arg)
			}
			port = args[i+1]
			i++
			continue
		}
		if arg == "--project-dir" {
			if i+1 > n {
				return fmt.Errorf("%s requires arg", arg)
			}
			projectDir = args[i+1]
			i++
			continue
		}

		if arg == "--include" {
			if i+1 > n {
				return fmt.Errorf("%s requires arg", arg)
			}
			include = append(include, args[i+1])
			i++
			continue
		}
		if arg == "--exclude" {
			if i+1 > n {
				return fmt.Errorf("%s requires arg", arg)
			}
			exclude = append(exclude, args[i+1])
			i++
			continue
		}
		if arg == "--build-arg" || arg == "--build-args" {
			if i+1 > n {
				return fmt.Errorf("%s requires arg", arg)
			}
			buildArgs = append(buildArgs, args[i+1])
			i++
			continue
		}
		if arg == "--diff-with" {
			if i+1 > n {
				return fmt.Errorf("%s requires arg", arg)
			}
			if args[i+1] == "" {
				return fmt.Errorf("%s requires value", arg)
			}
			diffWith = args[i+1]
			i++
			continue
		}
		if arg == "--full" {
			full = true
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			remain = append(remain, arg)
			continue
		}
		return fmt.Errorf("unknown flag: %s", arg)
	}

	if len(remain) == 0 {
		return fmt.Errorf("requires files")
	}

	ref := git.COMMIT_WORKING
	if diffWith == "" {
		diffWith = "origin/master"
	}

	opts := loadcov.LoadAllOptions{
		Dir:              projectDir,
		Args:             buildArgs,
		Profiles:         remain,
		Ref:              ref,
		DiffBase:         diffWith,
		Include:          include,
		Exclude:          exclude,
		OnlyChangedFiles: !full,
	}
	if cmd == "load" {
		data, err := loadcov.LoadAll(opts)
		if err != nil {
			return err
		}
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return err
		}
		fmt.Println(string(dataJSON))
		return nil
	}

	var actualPort int
	server := &http.ServeMux{}
	serve.RouteServer(server, "", func() int {
		return actualPort
	}, opts)

	autoIncrPort := true
	h, p := netutil.GetHostAndIP(bind, port)
	return netutil.ServePortHTTP(server, h, p, autoIncrPort, 500*time.Millisecond, func(port int) {
		actualPort = port
		url, extra := netutil.GetURLToOpen(h, port)
		netutil.PrintUrls(url, extra...)
		openURL(url)
	})
}

func openURL(url string) {
	openCmd := "open"
	if runtime.GOOS == "windows" {
		openCmd = "explorer"
	}
	cmd.Run(openCmd, url)
}
