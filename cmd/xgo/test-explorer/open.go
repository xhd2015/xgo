package test_explorer

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/netutil"
)

func setupOpenHandler(server *http.ServeMux) {
	server.HandleFunc("/openVscode", func(w http.ResponseWriter, r *http.Request) {
		handleOpenFile(w, r, func(file string, line int) error {
			if line > 0 {
				return cmd.Debug().Run("code", "--goto", fmt.Sprintf("%s:%d", file, line))
			}
			return cmd.Debug().Run("code", file)
		})
	})
	server.HandleFunc("/openGoland", func(w http.ResponseWriter, r *http.Request) {
		handleOpenFile(w, r, func(file string, line int) error {
			// see https://www.jetbrains.com/help/go/working-with-the-ide-features-from-command-line.html#standalone
			//  and https://www.jetbrains.com/help/go/opening-files-from-command-line.html#macos
			// open -na "GoLand.app" --args "$@"
			var args []string
			if line > 0 {
				args = append(args, "--line", strconv.Itoa(line))
			}
			args = append(args, file)
			switch runtime.GOOS {
			case "darwin":
				return cmd.Debug().Run("open", append([]string{"-na", "GoLand.app", "--args"}, args...)...)
			case "linux":
				return cmd.Debug().Run("goland.sh", args...)
			case "windows":
				return cmd.Debug().Run("goland64.exe", args...)
			default:
				return fmt.Errorf("open goland not supported on %s", runtime.GOOS)
			}
		})
	})
}

func handleOpenFile(w http.ResponseWriter, r *http.Request, callback func(file string, line int) error) {
	netutil.SetCORSHeaders(w)
	netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
		q := r.URL.Query()
		file := q.Get("file")
		line := q.Get("line")

		lineNum, _ := strconv.Atoi(line)
		return nil, callback(file, lineNum)
	})
}
