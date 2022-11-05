package main

import (
	"io"
	"os"
	"path/filepath"

	formatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
)

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

	log.SetOutput(io.MultiWriter(os.Stderr, mainLogFile))
	log.SetLevel(log.InfoLevel)

	initTemplate()

	go httpServer(config.HttpListenAddr)
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

			_, err = kernel.c.AddJob(job.JobConfig.Cron, job)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(kernel.c.Entries()) == 0 {
		log.Warn("No jobs loaded.")
	} else {
		log.Infof("Loaded jobs count: %d", len(kernel.c.Entries()))
	}

	return nil
}

func main() {
	kernel = NewKernel()

	s, err := service.New(kernel, kernel.svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		_ = logger.Error(err)
	}
}
