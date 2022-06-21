package main

const (
	defaultConfigFilePath = "gron.conf"

	defaultOnSuccessMessageFmt = "Job {{.JobName}} finished."
	defaultOnErrorMessageFmt   = "Job {{.JobName}} failed:\n\n{{.Error}}"
)
