package main

import (
	"fmt"
	"os"
)

type logFile struct{ file *os.File }

func (lf *logFile) Printf(format string, v ...interface{}) {
	fmt.Fprintf(lf.file, format, v...)
}

func (lf *logFile) Println(v ...interface{}) {
	fmt.Fprintln(lf.file, v...)
}

var mainLogFile *os.File

func initLogFile() error {
	var err error
	mainLogFile, err = os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	return err
}
