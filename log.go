package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var logFile *os.File

func initLogFile() {
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
