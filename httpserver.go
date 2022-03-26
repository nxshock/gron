package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func httpServer(listenAddress string) {
	http.HandleFunc("/", handler)
	http.HandleFunc("/shutdown", handleShutdown)
	http.HandleFunc("/start", handleForceStart)
	log.WithField("job", "http_server").Fatal(http.ListenAndServe(listenAddress, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	currentRunningJobsMutex.RLock()
	buf := new(bytes.Buffer)
	jobEntries := c.Entries()
	var jobs []*Job
	for _, v := range jobEntries {
		jobs = append(jobs, v.Job.(*Job))
	}
	indexTemplate.ExecuteTemplate(buf, "index", jobs)
	currentRunningJobsMutex.RUnlock()

	buf.WriteTo(w)
}

func handleForceStart(w http.ResponseWriter, r *http.Request) {
	jobName := r.FormValue("jobName")
	if jobName == "" {
		http.Error(w, "job name is not specified", http.StatusBadRequest)
		return
	}

	jobEntries := c.Entries()

	for _, jobEntry := range jobEntries {
		job := jobEntry.Job.(*Job)
		if job.FileName == jobName {
			log.WithField("job", "http_server").Printf("forced start %s", job.FileName)
			go job.Run()
			time.Sleep(time.Second / 4) // wait some time for job start
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}

	http.Error(w, fmt.Sprintf("there is no job with name %s", jobName), http.StatusBadRequest)
}

func handleShutdown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application terminated.\n"))

	go func() {
		time.Sleep(time.Second)
		log.WithField("job", "http_server").Infoln("Shutdown requested")
		os.Exit(0)
	}()
}
