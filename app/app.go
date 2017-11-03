// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/timshannon/townsourced/app/email"
	"github.com/timshannon/townsourced/data"
)

var (
	devMode    = false
	testMode   = false
	httpClient *http.Client
	baseURL    string
)

// Config are the config values for
// starting up the application layer
type Config struct {
	HTTPClientTimeout string `json:"httpClientTimeout"`
	DevMode           bool   `json:"-"`
	TestMode          bool   `json:"-"`
	TaskQueueSize     uint   `json:"taskQueueSize"`
	TaskPollTime      string `json:"taskPollTime"`
}

// DefaultConfig returns the default configuration for the app layer
func DefaultConfig() *Config {
	return &Config{
		HTTPClientTimeout: "30s",
		TaskQueueSize:     100,
		TaskPollTime:      "1m",
	}
}

// Init initializes the application layer
func Init(cfg *Config, hostname, siteURL, runningDir string) error {
	devMode = cfg.DevMode
	timeout, err := time.ParseDuration(cfg.HTTPClientTimeout)
	if err != nil {
		return fmt.Errorf("Error parsing HTTPClientTimeout duration: %s", err)
	}

	urlFormat, err := url.Parse(siteURL)
	if err != nil {
		return fmt.Errorf("Error parsing siteURL: %s", err)
	}

	//ensure a clean baseURL
	urlFormat.Path = ""
	urlFormat.Fragment = ""
	baseURL = urlFormat.String()

	httpClient = &http.Client{
		Timeout: timeout,
	}

	taskPoll, err := time.ParseDuration(cfg.TaskPollTime)
	if err != nil {
		return err
	}

	err = ensureAnnouncementTown()
	if err != nil {
		return err
	}

	initTasks(data.Key(hostname), cfg.TaskQueueSize, taskPoll)

	err = email.Init(cfg.TestMode)

	if err != nil {
		return err
	}

	initMessages(runningDir)

	return nil
}
