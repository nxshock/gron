package main

import (
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	formatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var c *cron.Cron

func init() {
	err := initConfig()
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(config.LogFilesPath, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	err = initLogFile()
	if err != nil {
		log.Fatalln(err)
	}

	log.SetFormatter(&formatter.Formatter{
		TimestampFormat: config.TimeFormat,
		HideKeys:        true,
		NoColors:        true,
		TrimMessages:    true})

	multiWriter := io.MultiWriter(os.Stderr, mainLogFile)
	log.SetOutput(multiWriter)

	log.SetLevel(log.InfoLevel)

	initTemplate()

	go httpServer(config.HttpListenAddr)

	c = cron.New()
}

func initJobs() error {
	log := log.WithField("job", "core")

	log.Infoln("Reading jobs...")
	err := filepath.Walk(config.JobConfigsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(info.Name()) == ".conf" {
			job, err := readJob(path)
			if err != nil {
				return err
			}

			_, err = c.AddJob(job.JobConfig.Cron, job)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(c.Entries()) == 0 {
		log.Warn("No jobs loaded.")
	} else {
		log.Infof("Loaded jobs count: %d", len(c.Entries()))
	}

	return nil
}

func main() {
	log := log.WithField("job", "core")

	log.Info("Started.")

	err := initJobs()
	if err != nil {
		log.Fatalln(err)
	}

	c.Start()

	intChan := make(chan os.Signal)
	signal.Notify(intChan, syscall.SIGTERM)
	<-intChan

	log.Info("Got stop signal.")

	err = mainLogFile.Close()
	if err != nil {
		log.Fatal(err)
	}
}
