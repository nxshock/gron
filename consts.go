package main

import formatter "github.com/antonfisher/nested-logrus-formatter"

const (
	defaultConfigFilePath = "gron.conf"
)

var (
	logFormat = &formatter.Formatter{
		TimestampFormat: config.TimeFormat,
		HideKeys:        true,
		NoColors:        true,
		TrimMessages:    true}
)
