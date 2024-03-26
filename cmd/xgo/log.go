package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var logDebugFile *os.File

func setupDebugLog(logDebugOption *string) (func(), error) {
	var logDebugFileName string
	if logDebugOption != nil {
		logDebugFileName = *logDebugOption
		if logDebugFileName == "" {
			logDebugFileName = "debug.log"
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
