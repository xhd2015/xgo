package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/support/osinfo"
)

func setupDebugLog(logDebugOption *string) (func(), error) {
	return config.SetupDebugLog(logDebugOption)
}

func logDebug(format string, args ...interface{}) {
	config.LogDebug(format, args...)
}

func logStartup() {
	logDebug("start: %v", __DEBUG_CMD_ARGS(os.Args))
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
