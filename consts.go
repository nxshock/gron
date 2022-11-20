package main

const (
	defaultConfigFileName = "gron.conf"

	defaultOnSuccessMessageFmt = "Job {{.JobName}} finished."
	defaultOnErrorMessageFmt   = "Job {{.JobName}} failed:\n\n{{.Error}}"
)
