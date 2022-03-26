package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var c *cron.Cron

func init() {
	err := os.MkdirAll(logFilesPath, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	initLogFile()

	log.SetFormatter(logFormat)
	//multiWriter := io.MultiWriter(os.Stderr, logFile)
	//log.SetOutput(multiWriter)
	log.SetOutput(logFile)
	log.SetLevel(log.InfoLevel)

	initTemplate()

	go httpServer(listenAddress)

	c = cron.New()
}

func main() {
	log := log.WithField("job", "core")

	log.Info("started")

	err := filepath.Walk("jobs.d", func(path string, info os.FileInfo, err error) error {
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

			_, err = c.AddJob(job.Cron, job)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}

	if len(c.Entries()) == 0 {
		log.Fatal("no jobs loaded")
	}

	log.Infof("loaded jobs count: %d", len(c.Entries()))

	c.Start()

	intChan := make(chan os.Signal)
	signal.Notify(intChan, syscall.SIGTERM)
	<-intChan

	log.Info("got stop signal")

	err = logFile.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
}
