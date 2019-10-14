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

package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"sync"
	"tryffel.net/pkg/jellycli/api"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/player"
	"tryffel.net/pkg/jellycli/ui"
	"tryffel.net/pkg/jellycli/ui/controller"
)

func main() {
	logFile := setLogging()
	defer logFile.Close()
	logrus.Info("Starting jellycli")

	conf, err := config.NewSecretStore()
	if err != nil {
		fmt.Println("Failed to start application:", err)
		os.Exit(1)
	}

	host, err := conf.EnsureKey("jellyfin_host")
	if err != nil {
		fmt.Printf("Failed to get jellyfin host: %v", err)
	}

	logrus.Info("Connecting to ", host)
	client, err := api.NewApi(host)
	if err != nil {
		logrus.Errorf("Failed to initialize api: %v", err)
		os.Exit(1)
	}

	token, err := conf.GetKey("token")
	if token == "" {
		username, err := config.ReadUserInput("username", false)
		if err != nil {
			fmt.Printf("failed read username: %v", err)
			os.Exit(1)
		}

		password, err := config.ReadUserInput("password", true)
		if err != nil {
			fmt.Printf("failed to read password: %v", err)
			os.Exit(1)
		}

		err = client.Login(username, password)
		if err == nil && client.IsLoggedIn() {
			err = conf.SetKey("token", client.Token())
			if err != nil {
				fmt.Printf("failed to store token: %v", err)
				os.Exit(1)
			}
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}

	} else {
		err = client.SetToken(token)
	}

	userid, err := conf.GetKey("userid")
	if userid != "" {
		client.SetUserId(userid)
	} else {
		if err != nil {
			logrus.Error(fmt.Errorf("failed to set wallet value: %v", err))
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Printf("failure in login: %v", err)
	}

	content := controller.NewContent(client)

	gui, err := ui.NewGui(content)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	exitCode := 0

	if !client.ConnectionOk() {
		os.Exit(1)
	}

	p, err := player.NewPlayer(client)
	if err != nil {
		logrus.Error("failed to start media player: %v", err)
		os.Exit(1)
	}
	gui.AssignChannels(p.StateChannel(), p.ActionChannel())
	err = gui.Start()
	if err != nil {
		logrus.Error("failed to start gui update task: %v", err)
		exitCode = 1
	}
	_ = content.Start()
	err = p.Start()
	if err != nil {
		logrus.Error("failed to start media player task: %v", err)
		exitCode = 1
	}
	//p.RefreshState()
	err = gui.Show()
	if err != nil && err != gocui.ErrQuit {
		logrus.Error("Gui error: %v", err)
		exitCode = 1
	}

	err = p.Stop()
	if err != nil {
		logrus.Error("failed to stop media player task: %v", err)
		exitCode = 1
	}
	err = gui.Stop()
	if err != nil {
		logrus.Error("failed to stop gui update task: %v", err)
		exitCode = 1
	}

	_ = content.Stop()

	logrus.Info("Stopping applicaton")
	os.Exit(exitCode)
}

func setLogging() *os.File {
	logrus.SetLevel(logrus.DebugLevel)
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
	file, err := os.OpenFile("jellycli.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0760))
	if err != nil {
		logrus.Error("failed to open log file: ", err.Error())
		return nil
	}

	logrus.SetOutput(file)
	return file
}
