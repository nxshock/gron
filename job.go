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

type JobConfig struct {
	Cron        string
	Command     string
	Description string
}

var globalMutex sync.RWMutex

func readJob(filePath string) (*Job, error) {
	var jobConfig JobConfig

	_, err := toml.DecodeFile(filePath, &jobConfig)
	if err != nil {
		return nil, err
	}

	command, params := parseCommand(jobConfig.Command)

	job := &Job{
		Name:        strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
		Cron:        jobConfig.Cron,
		Command:     command,
		Params:      params,
		FileName:    strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filepath.Base(filePath))),
		Description: jobConfig.Description}

	return job, nil
}

func (js *JobConfig) Write() {
	buf := new(bytes.Buffer)
	toml.NewEncoder(buf).Encode(*js)
	ioutil.WriteFile("job.conf", buf.Bytes(), 0644)
}

type Job struct {
	Name string // from filename

	Cron        string   // cron decription
	Command     string   // command for execution
	Params      []string // command params
	FileName    string   // short job name
	Description string   // job description

	// Fields for stats
	CurrentRunningCount   int
	LastStartTime         string
	LastEndTime           string
	LastExecutionDuration string
	LastError             string
}

func (j *Job) Run() {
	startTime := time.Now()

	globalMutex.Lock()
	j.CurrentRunningCount++
	j.LastStartTime = startTime.Format(config.TimeFormat)
	globalMutex.Unlock()

	jobLogFile, _ := os.OpenFile(filepath.Join(config.LogFilesPath, j.FileName+".txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	defer jobLogFile.Close()
	defer jobLogFile.WriteString("\n")

	l := log.New()
	l.SetOutput(jobLogFile)
	l.SetFormatter(logFormat)

	log.WithField("job", j.FileName).Info("started")
	l.Info("started")

	cmd := exec.Command(j.Command, j.Params...)
	cmd.Stdout = jobLogFile
	cmd.Stderr = jobLogFile

	err := cmd.Run()
	if err != nil {
		log.WithField("job", j.FileName).Error(err.Error())
		l.WithField("job", j.FileName).Error(err.Error())

		globalMutex.Lock()
		j.LastError = err.Error()
		globalMutex.Unlock()
	} else {
		globalMutex.Lock()
		j.LastError = ""
		globalMutex.Unlock()
	}

	endTime := time.Now()
	log.WithField("job", j.FileName).Infof("finished (%s)", endTime.Sub(startTime).Truncate(time.Second).String())
	l.Infof("finished (%s)", endTime.Sub(startTime).Truncate(time.Second).String())

	globalMutex.Lock()
	j.CurrentRunningCount--
	j.LastEndTime = endTime.Format(config.TimeFormat)
	j.LastExecutionDuration = endTime.Sub(startTime).Truncate(time.Second).String()
	globalMutex.Unlock()
}
