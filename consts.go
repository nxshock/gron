package main

const (
	programName = "gron"

	defaultConfigFileName = "gron.conf"

	defaultOnSuccessMessageFmt = "Job {{.JobName}} finished."
	defaultOnErrorMessageFmt   = "Job {{.JobName}} failed:\n\n{{.Error}}"
)
