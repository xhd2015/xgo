package trace

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/xhd2015/xgo/cmd/xgo/trace/render"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/netutil"
)

const help = `
Xgo tool trace visualize a generated trace file.

Usage:
    xgo tool trace [options] <file>

Options:
    -v, --version <version>  specify the version of the trace file, default is 1.0

Examples:
    xgo test -run TestSomething --strace ./   generate trace file
    xgo tool trace TestSomething.json         visualize a generated trace

See https://github.com/xhd2015/xgo for documentation.

`

func Main(args []string) {
	var files []string
	var port string
	var bind string

	n := len(args)

	var showHelp bool
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			files = append(files, args[i+1:]...)
			break
		}
		if arg == "-h" || arg == "--help" {
			showHelp = true
			break
		}
		if arg == "--port" {
			if i+1 >= n {
				fmt.Fprintf(os.Stderr, "--port requires arg\n")
				os.Exit(1)
			}
			port = args[i+1]
			i++
			continue
		} else if strings.HasPrefix(arg, "--port=") {
			port = strings.TrimPrefix(arg, "--port=")
			continue
		}
		// add --bind
		if arg == "--bind" {
			if i+1 >= n {
				fmt.Fprintf(os.Stderr, "--bind requires arg\n")
				os.Exit(1)
			}
			bind = args[i+1]
			i++
			continue
		} else if strings.HasPrefix(arg, "--bind=") {
			bind = strings.TrimPrefix(arg, "--bind=")
			continue
		}

		if !strings.HasPrefix(arg, "-") {
			files = append(files, arg)
			continue
		}
		fmt.Fprintf(os.Stderr, "unrecognized flag: %s\n", arg)
		os.Exit(1)
	}
	if showHelp {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return
	}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "requires file\n")
		os.Exit(1)
	}
	if len(files) != 1 {
		fmt.Fprintf(os.Stderr, "xgo tool trace requires exactly 1 file, given: %v", files)
		os.Exit(1)
	}
	file := files[0]
	if port == "" {
		port = os.Getenv("PORT")
	}
	if bind == "" {
		bind = "localhost"
	}
	err := serveFile(bind, port, file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}

func serveFile(bindStr string, portStr string, file string) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("expect a trace file, given dir: %s", file)
	}
	server := http.NewServeMux()
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				stack := debug.Stack()
				io.WriteString(w, fmt.Sprintf("<pre>panic: %v\n%s</pre>", e, stack))
			}
		}()
		var stack *render.Stack

		stacks, ok, err := render.ReadStacks(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, fmt.Sprintf("%v", err))
			return
		}
		if !ok {
			record, err := parseRecord(file)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, fmt.Sprintf("%v", err))
				return
			}
			stack = convert(record)
			stacks = []*render.Stack{stack}
		}
		w.Header().Set("Content-Type", "text/html")
		render.RenderStacks(stacks, file, w)
	})
	server.HandleFunc("/openVscodeFile", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		file := q.Get("file")
		if file == "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "no file\n")
			return
		}
		_, err := os.Stat(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}
		line := q.Get("line")

		fileLine := file
		if line != "" {
			fileLine = file + ":" + line
		}
		output, err := exec.Command("code", "--goto", fileLine).Output()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if exitErr, ok := err.(*exec.ExitError); ok {
				w.Write(exitErr.Stderr)
			} else {
				io.WriteString(w, err.Error())
			}
			return
		}
		w.Write(output)
	})

	host, port := netutil.GetHostAndIP(bindStr, portStr)
	autoIncrPort := true
	err = netutil.ServePortHTTP(server, host, port, autoIncrPort, 500*time.Millisecond, func(port int) {
		url, extra := netutil.GetURLToOpen(host, port)
		netutil.PrintUrls(url, extra...)
		openURL(url)
	})
	if err != nil {
		return err
	}
	return nil
}

func openURL(url string) {
	openCmd := "open"
	if runtime.GOOS == "windows" {
		openCmd = "explorer"
	}
	cmd.Run(openCmd, url)
}

func parseRecord(file string) (*RootExport, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var root *RootExport
	err = json.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}
	return root, nil
}

func parseStack(file string) (*render.Stack, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var stack *render.Stack
	err = json.Unmarshal(data, &stack)
	if err != nil {
		return nil, err
	}
	return stack, nil
}
