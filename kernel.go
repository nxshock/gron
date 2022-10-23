package main

import (
	"github.com/kardianos/service"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type Kernel struct {
	// windows service data
	svcConfig *service.Config

	// Other data
	c *cron.Cron
}

func NewKernel() *Kernel {
	svcConfig := &service.Config{
		Name:        "gron",
		DisplayName: "Gron Job Scheduler",
		Description: "Gron Job Scheduler.",
	}

	kernel = &Kernel{
		svcConfig: svcConfig,
		c:         cron.New()}

	return kernel
}

func (k *Kernel) Start(s service.Service) error {
	go func() {
		log.WithField("job", "core").Info("Started.")

		err := initJobs()
		if err != nil {
			log.Fatalln(err)
		}

		kernel.c.Start()
	}()

	return nil
}

func (k *Kernel) Stop(s service.Service) error {
	log.Info("Got stop signal.")

	err := mainLogFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

var kernel *Kernel
