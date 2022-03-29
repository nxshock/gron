package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

// JobConfig is a TOML representation of job
type JobConfig struct {
	Cron        string // cron decription
	Command     string // command for execution
	Description string // job description
}

type Job struct {
	Name string // from filename

	JobConfig

	// Fields for stats
	CurrentRunningCount   int
	LastStartTime         string
	LastEndTime           string
	LastExecutionDuration string
	LastError             string
	NextLaunch            string
}

var globalMutex sync.RWMutex

func readJob(filePath string) (*Job, error) {
	var jobConfig JobConfig

	_, err := toml.DecodeFile(filePath, &jobConfig)
	if err != nil {
		return nil, err
	}

	job := &Job{
		Name:      strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
		JobConfig: jobConfig}

	return job, nil
}

func (js *JobConfig) Write() {
	buf := new(bytes.Buffer)
	toml.NewEncoder(buf).Encode(*js)
	ioutil.WriteFile("job.conf", buf.Bytes(), 0644)
}

func (j *Job) CommandAndParams() (command string, params []string) {
	quoted := false
	items := strings.FieldsFunc(j.JobConfig.Command, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})
	for i := range items {
		items[i] = strings.Trim(items[i], `"`)
	}

	return items[0], items[1:]
}

func (j *Job) Run() {
	startTime := time.Now()

	globalMutex.Lock()
	j.CurrentRunningCount++
	j.LastStartTime = startTime.Format(config.TimeFormat)
	globalMutex.Unlock()

	jobLogFile, _ := os.OpenFile(filepath.Join(config.LogFilesPath, j.Name+".txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	defer jobLogFile.Close()
	defer jobLogFile.WriteString("\n")

	l := log.New()
	l.SetOutput(jobLogFile)
	l.SetFormatter(log.StandardLogger().Formatter)

	log.WithField("job", j.Name).Info("started")
	l.Info("started")

	command, params := j.CommandAndParams()

	cmd := exec.Command(command, params...)
	cmd.Stdout = jobLogFile
	cmd.Stderr = jobLogFile

	err := cmd.Run()
	if err != nil {
		log.WithField("job", j.Name).Error(err.Error())
		l.WithField("job", j.Name).Error(err.Error())

		globalMutex.Lock()
		j.LastError = err.Error()
		globalMutex.Unlock()
	} else {
		globalMutex.Lock()
		j.LastError = ""
		globalMutex.Unlock()
	}

	endTime := time.Now()
	log.WithField("job", j.Name).Infof("finished (%s)", endTime.Sub(startTime).Truncate(time.Second).String())
	l.Infof("finished (%s)", endTime.Sub(startTime).Truncate(time.Second).String())

	globalMutex.Lock()
	j.CurrentRunningCount--
	j.LastEndTime = endTime.Format(config.TimeFormat)
	j.LastExecutionDuration = endTime.Sub(startTime).Truncate(time.Second).String()
	globalMutex.Unlock()
}
