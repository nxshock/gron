package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
)

var config Config

type Config struct {
	TimeFormat     string `default:"02.01.2006 15:04"`
	JobConfigsPath string `default:"gron.d"`
	LogFilePath    string `default:"gron.log"` // core log file path
	LogFilesPath   string `default:"logs"`     // job log files path
	HttpListenAddr string `default:"127.0.0.1:9876"`

	HttpProxyAddr string // proxy address for local http client
}

func initConfig() error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	if len(os.Args) > 2 {
		return fmt.Errorf("Usage: %s [path to config]", filepath.Base(ex))
	}

	configFilePath := defaultConfigFilePath
	if len(os.Args) == 2 {
		configFilePath = os.Args[1]
	}

	_, err = toml.DecodeFile(configFilePath, &config)
	if err != nil {
		return err
	}

	// Set defaults
	if err := defaults.Set(&config); err != nil {
		return err
	}

	return nil
}
