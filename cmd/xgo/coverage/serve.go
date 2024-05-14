package coverage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/httputil"
	"github.com/xhd2015/xgo/support/osinfo"
)

func handleServe(args []string) error {
	// download binary from coverage-visualizer/xgo-tool-coverage-serve
	toolDir, err := getToolDir()
	if err != nil {
		return err
	}
	serveTool := filepath.Join(toolDir, "xgo-tool-coverage-serve") + osinfo.EXE_SUFFIX

	ok, err := checkAndFixFile(serveTool)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(serveTool), 0755)
	if err != nil {
		return err
	}

	serveToolLastCheck := serveTool + ".last-check"

	s, err := readFileOrEmpty(serveToolLastCheck)
	if err != nil {
		return err
	}
	var lastCheckTime time.Time
	if s != "" {
		lastCheckTime, _ = time.Parse(dateTime, s)
	}

	// if tool does not exist, download it

	if !ok {
		// download
		fmt.Fprintf(os.Stderr, "downloading coverage-visualizer/xgo-tool-coverage-serve\n")
		err := downloadTool(serveTool, serveToolLastCheck)
		if err != nil {
			return err
		}
	} else if lastCheckTime.IsZero() || time.Since(lastCheckTime) > 7*24*time.Hour {
		go func() {
			defer func() {
				if e := recover(); e != nil {
					fmt.Fprintf(os.Stderr, "checking update: panic %v\n", e)
				}
			}()
			err := downloadTool(serveTool, serveToolLastCheck)
			if err != nil {
				fmt.Fprintf(os.Stderr, "checking update: %v\n", err)
			}
		}()
	}

	serveArgs := []string{"serve"}
	var hasBuildArg bool
	for _, arg := range args {
		if arg == "--build-arg" {
			hasBuildArg = true
			break
		}
	}
	if !hasBuildArg {
		serveArgs = append(serveArgs, "--build-arg", "./")
	}
	serveArgs = append(serveArgs, args...)
	stripOut := &stripWriter{w: os.Stdout}
	defer stripOut.Close()
	stripErr := &stripWriter{w: os.Stderr}
	defer stripErr.Close()
	return cmd.New().Stdout(stripOut).Stderr(stripErr).
		Run(serveTool, serveArgs...)
}

const dateTime = "2006-01-02 15:04:05"

func downloadTool(serveTool string, recordFile string) error {
	goos, err := getGOOS()
	if err != nil {
		return err
	}
	goarch, err := getGOARCH()
	if err != nil {
		return err
	}
	downloadURL := fmt.Sprintf("https://github.com/xhd2015/coverage-visualizer/releases/download/xgo-tool-coverage-serve-v0.0.1/xgo-tool-coverage-serve-%s-%s", goos, goarch)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	err = httputil.DownloadFile(ctx, downloadURL, serveTool)
	if err != nil {
		return err
	}
	recordTime := time.Now().Format(dateTime)
	os.WriteFile(recordFile, []byte(recordTime), 0755)

	return chmodExec(serveTool)
}

// TODO: make executable on windows
func chmodExec(file string) error {
	return os.Chmod(file, 0755)
}

func getGOOS() (string, error) {
	goos := runtime.GOOS
	if goos != "" {
		return goos, nil
	}
	goos, err := cmd.Output("go", "env", "GOOS")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(goos), nil
}
func getGOARCH() (string, error) {
	goarch := runtime.GOARCH
	if goarch != "" {
		return goarch, nil
	}
	goarch, err := cmd.Output("go", "env", "GOARCH")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(goarch), nil
}

func readFileOrEmpty(file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func checkAndFixFile(file string) (bool, error) {
	statInfo, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if statInfo.IsDir() {
		err = os.RemoveAll(file)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func getToolDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".xgo", "tool"), nil
}

type stripWriter struct {
	buf []byte
	w   io.Writer
}

func (c *stripWriter) Write(p []byte) (int, error) {
	n := len(p)
	for i := 0; i < n; i++ {
		if p[i] != '\n' {
			c.buf = append(c.buf, p[i])
			continue
		}
		err := c.sendLine()
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (c *stripWriter) sendLine() error {
	str := string(c.buf)
	c.buf = nil
	for _, stripPair := range stripPairs {
		from := string(stripPair[0])
		to := string(stripPair[1])
		str = strings.ReplaceAll(str, from, to)
	}

	_, err := c.w.Write([]byte(str))
	if err != nil {
		return err
	}
	_, err = c.w.Write([]byte{'\n'})
	return err
}

var stripPairs = [][2][]byte{
	{
		[]byte{103, 105, 116, 46, 103, 97, 114, 101, 110, 97, 46, 99, 111, 109},
		[]byte{115, 111, 109, 101, 45, 103, 105, 116, 46, 99, 111, 109},
	},
	{
		[]byte{115, 104, 111, 112, 101, 101},
		[]byte{115, 111, 109, 101, 45, 99, 111, 114, 112},
	},
	{
		[]byte{108, 111, 97, 110, 45, 115, 101, 114, 118, 105, 99, 101},
		[]byte{115, 111, 109, 101, 45, 115, 101, 114, 118, 105, 99, 101},
	},
	{
		[]byte{99, 114, 101, 100, 105, 116, 95, 98, 97, 99, 107, 101, 110, 100},
		[]byte{115, 111, 109, 101, 45, 108, 118, 49},
	},
}

func (c *stripWriter) Close() error {
	if len(c.buf) > 0 {
		c.sendLine()
	}
	return nil
}
