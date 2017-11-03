// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"log"
	"net/http"

	"git.townsourced.com/townsourced/httprouter"
	"git.townsourced.com/townsourced/logrus"
)

// startSSLForward will start a server on port 80 which forwards
// to the ssl server on 443
func startSSLForwarder(c *Config) {
	forwardHandler := httprouter.New()

	forwardHandler.HandlerFunc("GET", "/*all", sslForwardHandler)
	forwardHandler.HandlerFunc("PUT", "/*all", sslForwardHandler)
	forwardHandler.HandlerFunc("POST", "/*all", sslForwardHandler)
	forwardHandler.HandlerFunc("DELETE", "/*all", sslForwardHandler)

	go func(cfg *Config) {
		server := &http.Server{
			Handler:        forwardHandler,
			ReadTimeout:    cfg.readTimeout,
			WriteTimeout:   cfg.writeTimeout,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
			ErrorLog:       log.New(logrus.StandardLogger().Writer(), "", log.LstdFlags),
		}

		server.Addr = ":http"
		err := server.ListenAndServe()
		if err != nil {
			logrus.WithField("error", err).Error("Error starting ssl forwarding server")
		}
	}(c)
}

func sslForwardHandler(w http.ResponseWriter, r *http.Request) {
	sslURL := *r.URL
	sslURL.Scheme = "https"
	sslURL.Host = r.Host

	http.Redirect(w, r, sslURL.String(), http.StatusMovedPermanently)
	return
}
