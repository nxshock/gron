package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WsConnections struct {
	connections map[*websocket.Conn]struct{}
	mutex       sync.Mutex
}

func (wc *WsConnections) Add(c *websocket.Conn) {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	wc.connections[c] = struct{}{}
}

func (wc *WsConnections) Delete(c *websocket.Conn) {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	delete(wc.connections, c)
}

func (wc *WsConnections) Send(message interface{}) {
	for conn := range wc.connections {
		go func(conn *websocket.Conn) { _ = conn.WriteJSON(message) }(conn)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsConnections = &WsConnections{
	connections: make(map[*websocket.Conn]struct{})}

func httpServer(listenAddress string) {
	if listenAddress == "none" {
		return
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/reloadJobs", handleReloadJobs)
	http.HandleFunc("/shutdown", handleShutdown)
	http.HandleFunc("/start", handleForceStart)
	http.HandleFunc("/details", handleDetails)
	http.HandleFunc("/ws", handleWebSocket)
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

	jobs := make(map[string][]*Job)
	for _, jobEntry := range kernel.c.Entries() {
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

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	wsConnections.Add(conn)
	defer wsConnections.Delete(conn)

	var startMessage struct {
		JobName string
	}

	for {
		err := conn.ReadJSON(&startMessage)
		if err != nil {
			log.Println(err)
			break
		}

		for _, jobEntry := range kernel.c.Entries() {
			job := jobEntry.Job.(*Job)
			if job.Name == startMessage.JobName {
				host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
				if err != nil {
					host = r.RemoteAddr
				}
				log.WithField("job", "http_server").Printf("Forced start %s from %s.", job.Name, host)
				go job.Run()
				break
			}
		}
	}
}

func handleForceStart(w http.ResponseWriter, r *http.Request) {
	jobName := r.FormValue("jobName")
	if jobName == "" {
		http.Error(w, "job name is not specified", http.StatusBadRequest)
		return
	}

	for _, jobEntry := range kernel.c.Entries() {
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
		_ = kernel.Stop(nil)
	}()
}

func handleReloadJobs(w http.ResponseWriter, r *http.Request) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	kernel.c.Stop()

	for _, entry := range kernel.c.Entries() {
		kernel.c.Remove(entry.ID)
	}

	err := initJobs()
	if err != nil {
		http.Error(w, fmt.Sprintf("reload jobs error: %v", err), http.StatusInternalServerError)
		return
	}

	kernel.c.Start()

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func handleDetails(w http.ResponseWriter, r *http.Request) {
	jobName := r.FormValue("jobName")
	if jobName == "" {
		http.Error(w, "job name is not specified", http.StatusBadRequest)
		return
	}

	jobEntries := kernel.c.Entries()

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
