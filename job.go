package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	formatter "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

// JobConfig is a TOML representation of job
type JobConfig struct {
	Type JobType

	Category string

	// JobType = Cmd
	Command    string // command for execution
	WorkingDir string // working directory

	// JobType = Sql
	Driver           string
	ConnectionString string
	SqlText          string

	// Other fields
	Cron                   string // cron decription
	Description            string // job description
	NumberOfRestartAttemts int
	RestartSec             int         // the time to sleep before restarting a job (seconds)
	RestartRule            RestartRule // Configures whether the job shall be restarted when the job process exits

	OnSuccessMessageFmt string // Success message format
	OnErrorMessageFmt   string // Error message format

	OnSuccessCmd string
	OnErrorCmd   string

	OnSuccessHttpPostUrl string
	OnErrorHttpPostUrl   string

	OnSuccessHttpGetUrl string
	OnErrorHttpGetUrl   string
}

type Job struct {
	Name string // from filename

	JobConfig JobConfig

	// Fields for stats
	Status                Status
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

	if jobConfig.Type <= 0 {
		return nil, errors.New("job type is not specified")
	}

	if jobConfig.Type > 2 {
		return nil, fmt.Errorf("unknown job type id: %v", int(jobConfig.Type)) // TODO: add job name to log
	}

	jobName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	jobName = strings.ReplaceAll(jobName, " ", "")
	jobName = strings.ReplaceAll(jobName, "_", "")
	jobName = strings.ReplaceAll(jobName, "#", "")

	job := &Job{
		Name:      jobName,
		Status:    Inactive,
		JobConfig: jobConfig}

	return job, nil
}

func splitCommandAndParams(s string) (command string, params []string) {
	quoted := false
	items := strings.FieldsFunc(s, func(r rune) bool {
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

// logEntry - logger which merged with main logger,
// jobLogFile - job log file with needs to be closed after job is done
func (j *Job) openAndMergeLog() (logEntry *log.Entry, jobLogFile *os.File) {
	jobLogFile, _ = os.OpenFile(filepath.Join(config.LogFilesPath, j.Name+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644) // TODO: handle error
	_, _ = jobLogFile.WriteString("\n")

	logWriter := io.MultiWriter(mainLogFile, jobLogFile)

	log := log.New()
	log.SetFormatter(&formatter.Formatter{
		TimestampFormat: config.TimeFormat,
		HideKeys:        true,
		NoColors:        true,
		TrimMessages:    true})
	log.SetOutput(logWriter)
	logEntry = log.WithField("job", j.Name)

	return logEntry, jobLogFile
}

func (j *Job) Run() {
	// TODO: переписать неоптимальный цикл
	for _, jobEntry := range kernel.c.Entries() {
		if jobEntry.Job.(*Job) != j {
			continue
		}

		j.NextLaunch = jobEntry.Next.Format(config.TimeFormat)
		break
	}

	log, jobLogFile := j.openAndMergeLog()
	defer jobLogFile.Close()

	for i := 0; i < j.JobConfig.NumberOfRestartAttemts+1; i++ {
		err := j.runTry(log, jobLogFile)
		if err == nil {
			break
		}

		if j.JobConfig.RestartRule == No || j.JobConfig.NumberOfRestartAttemts == 0 {
			break
		}

		if i == 0 {
			log.Printf("Job failed, restarting in %d seconds.", j.JobConfig.RestartSec)
			j.Status = Restarting
		} else if i < j.JobConfig.NumberOfRestartAttemts {
			j.Status = Restarting
			log.Printf("Retry attempt №%d of %d failed, restarting in %d seconds.", i, j.JobConfig.NumberOfRestartAttemts, j.JobConfig.RestartSec)
		} else {
			log.Printf("Retry attempt №%d of %d failed.", i, j.JobConfig.NumberOfRestartAttemts)
		}

		time.Sleep(time.Duration(j.JobConfig.RestartSec) * time.Second)
	}
}

func (j *Job) runTry(log *log.Entry, jobLogFile *os.File) error {
	log.Info("Started.")
	startTime := time.Now()

	globalMutex.Lock()
	j.CurrentRunningCount++
	j.Status = Running
	j.LastStartTime = startTime.Format(config.TimeFormat)
	globalMutex.Unlock()

	wsConnections.Send(j)

	var err error
	switch j.JobConfig.Type {
	case Cmd:
		err = j.runCmd(jobLogFile)
	case Sql:
		err = j.runSql(jobLogFile)
	}
	if err != nil {
		j.Status = Error
		log.Error(err.Error())

		globalMutex.Lock()
		j.LastError = err.Error()
		globalMutex.Unlock()
	} else {
		j.Status = Inactive
		globalMutex.Lock()
		j.LastError = ""
		globalMutex.Unlock()
	}
	go j.runFinishCallback(log, err)

	endTime := time.Now()
	log.Infof("Finished (%s).", endTime.Sub(startTime).Truncate(time.Second).String())

	globalMutex.Lock()
	j.CurrentRunningCount--
	j.LastEndTime = endTime.Format(config.TimeFormat)
	j.LastExecutionDuration = endTime.Sub(startTime).Truncate(time.Second).String()
	globalMutex.Unlock()

	wsConnections.Send(j)

	return err
}

func (j *Job) runCmd(jobLogFile *os.File) error {
	command, params := splitCommandAndParams(j.JobConfig.Command)

	cmd := exec.Command(command, params...)
	cmd.Stdout = jobLogFile
	cmd.Stderr = jobLogFile
	cmd.Dir = j.JobConfig.WorkingDir

	return cmd.Run()
}

func (j *Job) runSql(jobLogFile *os.File) error {
	var (
		db  *sql.DB
		err error
	)

	switch j.JobConfig.Driver {
	case "mssql", "sqlserver":
		db, err = openMsSqlDb(j.JobConfig.ConnectionString, jobLogFile)
	default:
		db, err = sql.Open(j.JobConfig.Driver, j.JobConfig.ConnectionString)
	}
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), defaultDbTimeout)
	defer cancel()

	_, err = db.ExecContext(ctx, j.JobConfig.SqlText)
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) runFinishCallback(log *log.Entry, jobErr error) error {
	s := j.JobConfig.OnSuccessCmd

	// Fill variables
	errStr := ""
	if jobErr != nil {
		errStr = jobErr.Error()
	}
	if jobErr != nil {
		s = format(s, struct{ Error string }{Error: errStr})
	}

	if jobErr == nil && j.JobConfig.OnSuccessCmd != "" {
		cmd, params := splitCommandAndParams(s)
		return runSimpleCmd(cmd, params...)
	}

	if jobErr == nil && j.JobConfig.OnSuccessHttpPostUrl != "" {
		err := httpPost(j.JobConfig.OnSuccessHttpPostUrl, j.successMessage())
		if err != nil {
			log.Error(err)
		}
	}

	if jobErr == nil && j.JobConfig.OnSuccessHttpGetUrl != "" {
		err := httpGet(j.JobConfig.OnSuccessHttpPostUrl, j.Name, j.successMessage())
		if err != nil {
			log.Error(err)
		}
	}

	if jobErr != nil && j.JobConfig.OnErrorCmd != "" {
		cmd, params := splitCommandAndParams(s)
		return runSimpleCmd(cmd, params...)
	}

	if jobErr != nil && j.JobConfig.OnErrorHttpPostUrl != "" {
		err := httpPost(j.JobConfig.OnErrorHttpPostUrl, j.errorMessage(jobErr))
		if err != nil {
			log.Error(err)
		}
	}

	if jobErr != nil && j.JobConfig.OnErrorHttpGetUrl != "" {
		err := httpGet(j.JobConfig.OnErrorHttpGetUrl, j.Name, jobErr.Error())
		if err != nil {
			log.Errorf("OnErrorHttpGetUrl error: %v", err) // TODO: сделать формат сообщения по стандарту
		}
	}

	return nil
}

func (j *Job) successMessage() string {
	s := j.JobConfig.OnSuccessMessageFmt
	if s == "" {
		s = defaultOnSuccessMessageFmt
	}

	v := struct {
		JobName string
	}{
		JobName: j.Name}

	return format(s, v)
}

func (j *Job) errorMessage(err error) string {
	s := j.JobConfig.OnErrorMessageFmt
	if s == "" {
		s = defaultOnErrorMessageFmt
	}

	v := struct {
		JobName string
		Error   string
	}{
		JobName: j.Name,
		Error:   err.Error()}

	return format(s, v)
}

func runSimpleCmd(command string, args ...string) error {
	return exec.Command(command, args...).Run()
}
