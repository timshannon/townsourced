// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"git.townsourced.com/townsourced/config"
	"git.townsourced.com/townsourced/logrus"
	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/web"
)

var (
	hostname      = ""
	flagDevMode   = false
	flagDemoMode  = false
	flagDir       = "."
	flagZopfli    = false
	flagSubdomain = ""
)

func init() {
	flag.BoolVar(&flagDevMode, "dev", false, "Dev mode prints out logs to console, and rebuilds templates on each "+
		"call, so they are updated on the fly.")
	flag.BoolVar(&flagDemoMode, "demo", false, "Demo mode runs townsourced with a different login page and "+
		"doesn't allow new signups.")
	flag.BoolVar(&flagZopfli, "zopfli", false, "Use zopfli compression instead of gzip.  It makes the townsourced "+
		"server startup slower, but creates smaller file sizes for static assets.")
	flag.StringVar(&flagDir, "dir", ".", "Dir sets the directory where server files will be served from.")
	flag.StringVar(&flagSubdomain, "subdomain", "", "Only works in dev mode, forces townsourced to a specific subdomain.")

	go func() {
		//Capture program shutdown, to make sure everything shuts down nicely
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for sig := range c {
			if sig == os.Interrupt {
				app.Halt("Townsourced Web Server %s shutting down", hostname)
			}
		}
	}()
}

func main() {
	flag.Parse()
	var err error

	hostname, err = os.Hostname()
	if err != nil {
		app.Halt("Error retrieving hostname: %s", err)
	}

	settingPaths := config.StandardFileLocations("townsourced/settings.json")
	fmt.Println("This townsourced webserver will use settings files in the following locations (in order of priority):")
	for i := range settingPaths {
		fmt.Println("\t", settingPaths[i])
	}
	cfg, err := config.LoadOrCreate(settingPaths...)
	if err != nil {
		app.Halt(err.Error())
	}

	fmt.Printf("This webserver is currently using the file %s for settings.\n", cfg.FileName())

	err = os.Chdir(flagDir)
	if err != nil {
		app.Halt("Error changing dir to  %s: %s", flagDir, err)
	}

	webCfg := web.DefaultConfig()

	err = cfg.ValueToType("web", webCfg)
	if err != nil {
		app.Halt("Error reading web config values: %s", err.Error())
	}

	webCfg.DevMode = flagDevMode
	webCfg.DemoMode = flagDemoMode
	webCfg.Zopfli = flagZopfli
	webCfg.SubDomain = flagSubdomain

	dataCfg := data.DefaultConfig()
	err = cfg.ValueToType("data", dataCfg)
	if err != nil {
		app.Halt("Error reading data config values: %s", err.Error())
	}

	dataCfg.DevMode = flagDevMode

	err = cfg.Write()
	if err != nil {
		app.Halt("Error writting config file to %s. Error: %s", cfg.FileName(), err)
	}

	fmt.Printf("Townsourced Web Server %s starting up...\n", hostname)

	logrus.AddHook(&app.LogHook{})
	if flagDevMode {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	appCfg := app.DefaultConfig()

	err = cfg.ValueToType("app", appCfg)
	if err != nil {
		app.Halt("Error reading app config values: %s", err.Error())
	}
	appCfg.DevMode = flagDevMode

	err = data.Init(dataCfg)
	if err != nil {
		log.Fatalf("Error initializing townsourced data layer: %s", err.Error())
	}

	err = app.Init(appCfg, hostname, webCfg.Address, ".")
	if err != nil {
		log.Fatalf("Error initializing townsourced application layer: %s", err.Error())
	}

	err = web.StartServer(webCfg)
	if err != nil {
		app.Halt("Error Starting townsourced web server: %s", err.Error())
	}
}
