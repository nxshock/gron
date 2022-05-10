package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadJob(t *testing.T) {
	expectedJob := &Job{
		Name: "job",
		JobConfig: JobConfig{
			Type:                   Cmd,
			Cron:                   "* * * * *",
			Command:                `command "param1 param1" param2`,
			Description:            "comment",
			NumberOfRestartAttemts: 3,
			RestartSec:             5,
			RestartRule:            OnError}}

	job, err := readJob("tests/job.conf")
	assert.NoError(t, err)

	assert.Equal(t, expectedJob, job)
}
