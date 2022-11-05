package main

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

func createHttpClient() (*http.Client, error) {
	httpClient := new(http.Client)

	if config.HttpProxyAddr != "" {
		proxyURL, err := url.Parse(config.HttpProxyAddr)
		if err != nil {
			return nil, err
		}

		httpClient.Transport = &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}

	return httpClient, nil
}

func httpPost(addr string, text string) error {
	httpClient, err := createHttpClient()
	if err != nil {
		return err
	}

	resp, err := httpClient.Post(addr, "text/plain", strings.NewReader(text))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func httpGet(addrFmt, jobName, text string) error {
	httpClient, err := createHttpClient()
	if err != nil {
		return err
	}

	v := struct {
		JobName string
		Error   string
	}{
		JobName: url.QueryEscape(jobName),
		Error:   url.QueryEscape(text)}

	urlStr := format(addrFmt, v)
	log.Println(urlStr)

	resp, err := httpClient.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
