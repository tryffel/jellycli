/*
 * Copyright 2019 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package main contains jellycli executable and bootstraps application.
// Jellycli is a terminal application for playing music from Jellyfin server.
package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"io"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/api/jellyfin"
	"tryffel.net/go/jellycli/api/subsonic"
	"tryffel.net/go/jellycli/config"
	mpris2 "tryffel.net/go/jellycli/mpris"
	"tryffel.net/go/jellycli/player"
	"tryffel.net/go/jellycli/task"
	"tryffel.net/go/jellycli/ui"
)

// does config file need to be saved
var configChanged = false
var showGui = true

func main() {
	run()
}

// Application is the root struct for interactive player
type Application struct {
	conf        *config.Config
	api         api.MediaServer
	gui         *ui.Gui
	player      *player.Player
	mpris       *mpris2.MediaController
	mprisPlayer *mpris2.Player
	logfile     *os.File
	logFileName string
}

//NewApplication instantiates new player
func NewApplication(configFile string) (*Application, error) {
	var err error
	a := &Application{}

	err = a.initConfig(configFile)
	if err != nil {
		return a, err
	}

	config.ConfigFile = a.conf.ConfigFile()
	config.PageSize = a.conf.Player.PageSize
	config.LimitRecentlyPlayed = a.conf.Player.LimitRecentlyPlayed
	config.AudioBufferPeriod = time.Millisecond * time.Duration(a.conf.Player.AudioBufferingMs)

	a.logfile = setLogging(a.conf)

	// write to both log file and stdout for startup, in case there are any errors that prevent gui
	writer := io.MultiWriter(a.logfile, os.Stdout)
	logrus.SetOutput(writer)

	logrus.Infof("############# %s v%s ############", config.AppName, config.Version)
	err = a.initApi()
	if err != nil {
		return a, err
	}
	err = config.SaveConfig(a.conf)
	if err != nil {
		logrus.Errorf("save config file: %v", err)
	}

	config.AppConfig = a.conf
	err = a.initApplication()
	return a, err
}

func (a *Application) Start() error {
	if a.conf.Player.EnableRemoteControl {
		remoteController, ok := a.api.(api.RemoteController)
		if ok {
			logrus.Debug("Enable remote control")
			remoteController.SetPlayer(a.player)
			remoteController.SetQueue(a.player)
		}
	}
	var err error
	tasks := []task.Tasker{a.player, a.api}

	for _, v := range tasks {
		err = v.Start()
		if err != nil {
			return fmt.Errorf("failed to start tasks: %v", err)
		}
	}

	if showGui {
		logrus.SetOutput(a.logfile)
		go a.stopOnSignal()
		return a.gui.Start()
	} else {
		a.stopOnSignal()
	}
	return nil
}

func (a *Application) Stop() error {
	logrus.Info("Stopping application")
	tasks := []task.Tasker{a.player, a.api}
	var err error
	var hasError bool
	for _, v := range tasks {
		err = v.Stop()
		if err != nil {
			logrus.Error(err)
			hasError = true
		}
	}
	if showGui {
		a.gui.Stop()
	}

	if err != nil || hasError {
		logrus.Errorf("stop application: %v", err)
		err = nil
	}

	if a.logfile != nil {
		err = a.logfile.Close()
		if err != nil {
			err = fmt.Errorf("close log file: %v", err)
		}
	}

	logrus.SetOutput(os.Stdout)
	return err
}

func (a *Application) stopOnSignal() {
	<-catchSignals()
	a.Stop()
}

func (a *Application) initConfig(configFile string) error {
	var err error
	a.conf, err = config.ReadConfigFile(configFile)
	return err
}

func (a *Application) initApi() error {
	var err error
	switch strings.ToLower(a.conf.Player.Server) {
	case "jellyfin":
		a.api, err = jellyfin.NewJellyfin(&a.conf.Jellyfin, &config.StdinConfigProvider{})
		a.conf.Player.Server = "jellyfin"
	case "subsonic":
		a.api, err = subsonic.NewSubsonic(&a.conf.Subsonic, &config.StdinConfigProvider{})
		a.conf.Player.Server = "subsonic"
	default:
		return fmt.Errorf("unsupported backend: '%s'", a.conf.Player.Server)
	}
	if err != nil {
		return fmt.Errorf("api init: %v", err)
	}
	if err := a.api.ConnectionOk(); err != nil {
		return fmt.Errorf("no connection to server: %v", err)
	}

	conf := a.api.GetConfig()
	if a.conf.Player.Server == "jellyfin" {
		jfConfig, ok := conf.(*config.Jellyfin)
		if ok {
			a.conf.Jellyfin = *jfConfig
		}
	}
	if a.conf.Player.Server == "subsonic" {
		subConfig, ok := conf.(*config.Subsonic)
		if ok {
			a.conf.Subsonic = *subConfig
		}
	}
	return nil
}

func (a *Application) initApplication() error {
	var err error
	a.player, err = player.NewPlayer(a.api)
	if err != nil {
		return fmt.Errorf("create player: %v", err)
	}
	a.gui = ui.NewUi(a.player)

	a.mpris, err = mpris2.NewController(a.player.Audio)
	if err != nil {
		if strings.Contains(err.Error(), "dbus-launch") {
			logrus.Warningf("Dbus disabled: %v", err)
		} else {
			return fmt.Errorf("initialize dbus connection: %v", err)
		}
	} else {
		a.mprisPlayer = &mpris2.Player{
			MediaController: a.mpris,
		}
		a.player.AddStatusCallback(a.mprisPlayer.UpdateStatus)
	}
	return nil
}

func setLogging(conf *config.Config) *os.File {
	level, err := logrus.ParseLevel(conf.Player.LogLevel)
	if err != nil {
		logrus.Errorf("parse log level: %v", err)
		return nil
	}

	logrus.SetLevel(level)
	format := &prefixed.TextFormatter{
		ForceColors:      false,
		DisableColors:    true,
		ForceFormatting:  true,
		DisableTimestamp: false,
		DisableUppercase: false,
		FullTimestamp:    true,
		TimestampFormat:  "",
		DisableSorting:   false,
		QuoteEmptyFields: false,
		QuoteCharacter:   "'",
		SpacePadding:     0,
		Once:             sync.Once{},
	}
	logrus.SetFormatter(format)
	dir := os.TempDir()
	file := path.Join(dir, fmt.Sprintf("%s.log", strings.ToLower(config.AppName)))
	fd, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0760))
	if err != nil {
		logrus.Error("failed to open log fd: ", err.Error())
		return nil
	}

	config.LogFile = file
	logrus.SetOutput(fd)
	return fd
}

func catchSignals() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGINT,
		syscall.SIGTERM)
	return c
}

func run() {
	showVersion := flag.Bool("version", false, "Show version")
	configFile := flag.String("config", "",
		"Use external configuration file. file must be yaml-formatted")
	help := flag.Bool("help", false, "Show help page")
	gui := flag.Bool("no-gui", false, "Disable gui")
	flag.Parse()

	if *gui {
		showGui = false
	}

	if *showVersion {
		println(config.AppNameVersion())
	} else if *help {
		text := "Jellycli, a terminal music player for Jellyfin\n\nUsage:"
		println(text)
		flag.PrintDefaults()
	} else {
		app, err := NewApplication(*configFile)
		if err != nil {
			logrus.Fatal(err)
			fmt.Println(err)
			os.Exit(1)
		}

		startErr := app.Start()
		if startErr != nil {
			logrus.Errorf("Failed to start application: %v", startErr)
		}
		stopErr := app.Stop()
		if startErr == nil && stopErr == nil {
			os.Exit(0)
		}

		os.Exit(1)
	}
}
