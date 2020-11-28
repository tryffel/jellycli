/*
 * Copyright 2020 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/api/jellyfin"
	"tryffel.net/go/jellycli/api/subsonic"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/mpris"
	"tryffel.net/go/jellycli/player"
	"tryffel.net/go/jellycli/task"
	"tryffel.net/go/jellycli/ui"
)

type app struct {
	server      api.MediaServer
	gui         *ui.Gui
	player      *player.Player
	mpris       *mpris.MediaController
	mprisPlayer *mpris.Player
	logfile     *os.File
}

var disableGui = false

func initApplication() (*app, error) {

	if viper.GetBool("player_nogui") {
		disableGui = true
	}

	logFile, err := initLogging()
	if err != nil {
		logrus.Fatalf("init logging: %v", err)
	}

	a := &app{}
	a.logfile = logFile
	// write to both log file and stdout for startup, in case there are any errors that prevent gui
	writer := io.MultiWriter(a.logfile, os.Stdout)
	logrus.SetOutput(writer)

	logrus.Infof("############# %s v%s ############", config.AppName, config.Version)

	err = a.initServerConnection()
	if err != nil {
		logrus.Fatalf("connect to server: %v", err)
	}

	err = a.initApp()
	if err != nil {
		logrus.Fatalf("init application: %v", err)
	}

	a.initGui()
	err = config.SaveConfig()
	if err != nil {
		logrus.Errorf("save config file: %v", err)
	}

	a.run()
	stopErr := a.stop()
	if stopErr != nil {
		logrus.Fatal(stopErr)

		if stopErr == nil {
			os.Exit(0)
		}
	}

	return a, nil
}

func (a *app) initServerConnection() error {
	var err error
	switch strings.ToLower(config.AppConfig.Player.Server) {
	case "jellyfin":
		a.server, err = jellyfin.NewJellyfin(&config.AppConfig.Jellyfin, &config.ViperStdConfigProvider{})
	case "subsonic":
		a.server, err = subsonic.NewSubsonic(&config.AppConfig.Subsonic, &config.ViperStdConfigProvider{})
	default:
		return fmt.Errorf("unsupported backend: '%s'", config.AppConfig.Player.Server)
	}
	if err != nil {
		return fmt.Errorf("api init: %v", err)
	}
	if err := a.server.ConnectionOk(); err != nil {
		return fmt.Errorf("no connection to server: %v", err)
	}

	conf := a.server.GetConfig()
	if config.AppConfig.Player.Server == "jellyfin" {
		jfConfig, ok := conf.(*config.Jellyfin)
		if ok {
			config.AppConfig.Jellyfin = *jfConfig
		}
	} else if config.AppConfig.Player.Server == "subsonic" {
		subConfig, ok := conf.(*config.Subsonic)
		if ok {
			config.AppConfig.Subsonic = *subConfig
		}
	}
	return nil

}

func (a *app) initGui() {
	if !disableGui {
		a.gui = ui.NewUi(a.player)
	}
}

func (a *app) initApp() error {
	var err error
	a.player, err = player.NewPlayer(a.server)
	if err != nil {
		return fmt.Errorf("create player: %v", err)
	}
	a.mpris, err = mpris.NewController(a.player.Audio)
	if err != nil {
		if strings.Contains(err.Error(), "dbus-launch") {
			logrus.Warningf("Dbus disabled: %v", err)
		} else {
			return fmt.Errorf("initialize dbus connection: %v", err)
		}
	} else {
		a.mprisPlayer = &mpris.Player{
			MediaController: a.mpris,
		}
		a.player.AddStatusCallback(a.mprisPlayer.UpdateStatus)
	}
	return nil
}

func (a *app) run() {
	if config.AppConfig.Player.EnableRemoteControl {
		remoteController, ok := a.server.(api.RemoteController)
		if ok {
			logrus.Debug("Enable remote control")
			remoteController.SetPlayer(a.player)
			remoteController.SetQueue(a.player)
		}
	}
	var err error
	tasks := []task.Tasker{a.player, a.server}

	for _, v := range tasks {
		err = v.Start()
		if err != nil {
			logrus.Fatalf("start task: %v", err)
		}
	}

	if !disableGui {
		logrus.SetOutput(a.logfile)
		go a.stopOnSignal()
		err = a.gui.Start()
		if err != nil {
			logrus.Errorf("start gui: %v", err)
		}
	} else {
		a.stopOnSignal()
	}
}

func (a *app) stopOnSignal() {
	<-catchSignals()
	err := a.stop()
	if err != nil {
		logrus.Errorf("stop application: %v", err)
	}
}

func (a *app) stop() error {
	logrus.Info("Stopping application")
	tasks := []task.Tasker{a.player, a.server}
	var err error
	var hasError bool
	for _, v := range tasks {
		err = v.Stop()
		if err != nil {
			logrus.Error(err)
			hasError = true
		}
	}
	if !disableGui {
		a.gui.Stop()
	}

	if err != nil || hasError {
		logrus.Errorf("stop application: %v", err)
		err = nil
	}

	if a.logfile != nil {
		logrus.SetOutput(os.Stdout)
		err = a.logfile.Close()
		if err != nil {
			err = fmt.Errorf("close log file: %v", err)
		}
	}
	return err
}

func runApplication() {
	_, err := initApplication()
	if err != nil {
		logrus.Fatal(err)
	}
}

func catchSignals() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGINT,
		syscall.SIGTERM)
	return c
}
