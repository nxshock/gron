package main

import formatter "github.com/antonfisher/nested-logrus-formatter"

const (
	timeFormat    = "02.01.2006 15:04:05"
	logFileName   = "log.txt"
	logFilesPath  = "logs"
	listenAddress = "127.0.0.1:9876"
)

var (
	logFormat = &formatter.Formatter{
		TimestampFormat: timeFormat,
		HideKeys:        true,
		NoColors:        true,
		TrimMessages:    true}
)
