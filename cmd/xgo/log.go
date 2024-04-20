package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/osinfo"
)

var logDebugFile *os.File

func setupDebugLog(logDebugOption *string) (func(), error) {
	var logDebugFileName string
	if logDebugOption != nil {
		logDebugFileName = *logDebugOption
		if logDebugFileName == "" {
			// default to stderr
			logDebugFileName = "stderr"
		}
	}
	if logDebugFileName == "" {
		return nil, nil
	}
	return initLog(logDebugFileName)
}

func initLog(logDebugFileName string) (func(), error) {
	if logDebugFileName == "stdout" {
		logDebugFile = os.Stdout
		return nil, nil
	}
	if logDebugFileName == "stderr" {
		logDebugFile = os.Stderr
		return nil, nil
	}
	var err error
	logDebugFile, err = os.Create(logDebugFileName)
	if err != nil {
		return nil, fmt.Errorf("create log: %s %w", logDebugFileName, err)
	}
	return func() {
		logDebugFile.Close()
	}, nil
}

func logDebug(format string, args ...interface{}) {
	if logDebugFile == nil {
		return
	}
	fmt.Fprint(logDebugFile, time.Now().Format("2006-01-02 15:04:05"), " ")
	fmt.Fprintf(logDebugFile, format, args...)
	if !strings.HasSuffix(format, "\n") {
		fmt.Fprintln(logDebugFile)
	}
	logDebugFile.Sync()
}

func logStartup() {
	logDebug("start: %v", os.Args)
	logDebug("runtime.GOOS=%s", runtime.GOOS)
	logDebug("runtime.GOARCH=%s", runtime.GOARCH)
	logDebug("runtime.Version()=%s", runtime.Version())
	logDebug("runtime.GOROOT()=%s", runtime.GOROOT())
	logDebug("os exe suffix: %s", osinfo.EXE_SUFFIX)
	logDebug("os force copy unsym: %v", osinfo.FORCE_COPY_UNSYM)
}

// if [[ $verbose = true ]];then
//
//	    tail -fn1 "$shdir/compile.log" &
//	    trap "kill -9 $!" EXIT
//	fi
func tailLog(logFile string) {
	file, err := os.OpenFile(logFile, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open compile log: %v\n", err)
		return
	}
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "seek tail compile log: %v\n", err)
		return
	}
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			fmt.Fprintf(os.Stderr, "tail compile log: %v\n", err)
			return
		}
	}
}
