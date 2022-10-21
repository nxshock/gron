package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"sort"
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
	http.HandleFunc("/details", handleDetails)
	log.WithField("job", "http_server").Fatal(http.ListenAndServe(listenAddress, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		fs, err := fs.Sub(siteFS, "webui")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.FileServer(http.FS(fs)).ServeHTTP(w, r)
		return
	}

	globalMutex.RLock()
	buf := new(bytes.Buffer)
	jobEntries := c.Entries()

	jobs := make(map[string][]*Job)
	for _, jobEntry := range jobEntries {
		job := jobEntry.Job.(*Job)
		job.NextLaunch = jobEntry.Next.Format(config.TimeFormat)
		jobs[job.JobConfig.Category] = append(jobs[job.JobConfig.Category], job)
	}

	var keys []string
	for k, v := range jobs {
		keys = append(keys, k)

		sort.Slice(v, func(i, j int) bool {
			return v[i].Name < v[j].Name
		})
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	_ = templates.ExecuteTemplate(buf, "index.htm", struct {
		Categories []string
		Jobs       map[string][]*Job
	}{
		Categories: keys,
		Jobs:       jobs,
	})
	globalMutex.RUnlock()

	_, _ = buf.WriteTo(w)
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
		if job.Name == jobName {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			log.WithField("job", "http_server").Printf("Forced start %s from %s.", job.Name, host)
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
	_, _ = w.Write([]byte("Application terminated.\n"))

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

func handleDetails(w http.ResponseWriter, r *http.Request) {
	jobName := r.FormValue("jobName")
	if jobName == "" {
		http.Error(w, "job name is not specified", http.StatusBadRequest)
		return
	}

	jobEntries := c.Entries()

	for _, jobEntry := range jobEntries {
		job := jobEntry.Job.(*Job)
		if job.Name == jobName {
			err := templates.ExecuteTemplate(w, "details.htm", job)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}

	http.Error(w, fmt.Sprintf("there is no job with name %s", jobName), http.StatusBadRequest)
}
