package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func httpServer(listenAddress string) {
	if listenAddress == "none" {
		return
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/reloadJobs", handleReloadJobs)
	http.HandleFunc("/shutdown", handleShutdown)
	http.HandleFunc("/start", handleForceStart)
	log.WithField("job", "http_server").Fatal(http.ListenAndServe(listenAddress, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	globalMutex.RLock()
	buf := new(bytes.Buffer)
	jobEntries := c.Entries()
	var jobs []*Job
	for _, jobEntry := range jobEntries {
		job := jobEntry.Job.(*Job)
		job.NextLaunch = jobEntry.Next.Format(config.TimeFormat)
		jobs = append(jobs, job)
	}
	indexTemplate.ExecuteTemplate(buf, "index", jobs)
	globalMutex.RUnlock()

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
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			log.WithField("job", "http_server").Printf("forced start %s from %s", job.FileName, host)
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

func handleReloadJobs(w http.ResponseWriter, r *http.Request) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	c.Stop()

	for _, entry := range c.Entries() {
		c.Remove(entry.ID)
	}

	err := initJobs()
	if err != nil {
		http.Error(w, fmt.Sprintf("reload jobs error: %v", err), http.StatusInternalServerError)
		return
	}

	c.Start()

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
