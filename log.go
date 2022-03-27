package main

import (
	"os"
)

var logFile *os.File

func initLogFile() error {
	var err error
	logFile, err = os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	return err
}
