package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var compileLogFile *os.File

func initLog() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get config under home directory: %v", err)
	}

	xgoDir := filepath.Join(homeDir, ".xgo")
	logDir := filepath.Join(xgoDir, "log")
	compileLog := filepath.Join(logDir, "compile.log")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		return fmt.Errorf("create ~/.xgo/log: %w", err)
	}

	compileLogFile, err = os.OpenFile(compileLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("create ~/.xgo/log/compile.log: %w", err)
	}
	return nil
}

func logCompile(format string, args ...interface{}) {
	fmt.Fprint(compileLogFile, time.Now().Format("2006-01-02 15:04:05"), " ")
	fmt.Fprintf(compileLogFile, format, args...)
}
