package main

import (
	"os"

	"github.com/nxshock/logwriter"
)

type LogWriter struct{ w *logwriter.LogWriter }

func (lw *LogWriter) Printf(format string, v ...interface{}) {
	lw.w.Printf(format, v...)
}

func (lw *LogWriter) Println(v ...interface{}) {
	lw.w.Println(v...)
}

var mainLogFile *os.File

func initLogFile() error {
	var err error
	mainLogFile, err = os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	return err
}
