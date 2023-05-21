package main

import "time"

const (
	defaultConfigFileName = "gron.conf"

	defaultDbTimeout = 24 * time.Hour

	defaultOnSuccessMessageFmt = "Job {{.JobName}} finished."
	defaultOnErrorMessageFmt   = "Job {{.JobName}} failed:\n\n{{.Error}}"
)
