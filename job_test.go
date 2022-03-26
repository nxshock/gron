package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadJob(t *testing.T) {
	expectedJob := &Job{
		Name:        "job",
		Cron:        "* * * * *",
		Command:     "command",
		Params:      []string{"param1 param1", "param2"},
		FileName:    "job",
		Description: "comment"}

	job, err := readJob("tests/job.conf")
	assert.NoError(t, err)

	assert.Equal(t, expectedJob, job)
}
