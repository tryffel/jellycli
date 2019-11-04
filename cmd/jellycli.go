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
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"strconv"
	"sync"
	"tryffel.net/pkg/jellycli/api"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/controller"
	"tryffel.net/pkg/jellycli/player"
	"tryffel.net/pkg/jellycli/task"
	"tryffel.net/pkg/jellycli/ui"
)

func main() {

	app, err := NewApplication()
	//gui.AssignChannels(p.StateChannel(), p.ActionChannel())
	if err != nil {
		logrus.Fatal(err)
		fmt.Println(err)
		os.Exit(1)
	}

	startstartErr := app.Start()
	if startstartErr != nil {
		logrus.Error("Failed to start application: %v", startstartErr)
	}

	stopErr := app.Stop()

	if startstartErr == nil && stopErr == nil {
		os.Exit(0)
	}

	os.Exit(1)
}

// Application is the root struct for interactive player
type Application struct {
	secrets config.Secret
	api     *api.Api
	gui     *ui.Gui
	player  *player.Player
	content *controller.Content
	logfile *os.File
}

//NewApplication instantiates new player
func NewApplication() (*Application, error) {
	var err error
	a := &Application{}

	a.logfile = setLogging()
	err = a.initConfig()
	if err != nil {
		return a, err
	}
	logrus.Infof("############# %s v%s ############", config.AppName, config.Version)
	err = a.initApi()
	if err != nil {
		return a, err
	}
	err = a.login()
	if err != nil {
		return a, err
	}
	err = a.initApiView()
	if err != nil {
		return a, err
	}

	err = a.api.VerifyServerId()
	if err != nil {
		logrus.Fatal("api error: %v", err)
		os.Exit(1)
	}

	err = a.initApplication()
	return a, err
}

func (a *Application) Start() error {
	tasks := []task.Tasker{a.player, a.content, a.gui}
	var err error
	for _, v := range tasks {
		err = v.Start()
		if err != nil {
			return fmt.Errorf("failed to start tasks: %v", err)
		}
	}
	return nil
}

func (a *Application) Stop() error {
	logrus.Info("Stopping application")
	tasks := []task.Tasker{a.gui, a.player, a.content}
	var err error
	var hasError bool
	for _, v := range tasks {
		err = v.Stop()
		if err != nil {
			logrus.Error(err)
			hasError = true
		}
	}
	if a.logfile != nil {
		ferr := a.logfile.Close()
		logrus.Error("close log file: ", ferr)
		hasError = true
	}
	if err == nil && hasError == false {
		return nil
	}
	return fmt.Errorf("stop application: %v", err)
}

func (a *Application) initConfig() error {
	var err error
	a.secrets, err = config.NewSecretStore()
	if err != nil {
		return fmt.Errorf("wallet failed: %v", err)
	}
	return nil
}

func (a *Application) initApi() error {
	var err error
	host, err := a.secrets.EnsureKey("jellyfin_host")
	if err != nil {
		return fmt.Errorf("no jellyfin host provided: %v", err)
	}
	a.api, err = api.NewApi(host)
	if err != nil {
		return fmt.Errorf("api init: %v", err)
	}
	if !a.api.ConnectionOk() {
		return fmt.Errorf("no connection to server")
	}
	return nil
}

func (a *Application) login() error {
	token, _ := a.secrets.GetKey("token")
	if token == "" {
		username, err := config.ReadUserInput("username", false)
		if err != nil {
			return fmt.Errorf("failed read username: %v", err)
		}

		password, err := config.ReadUserInput("password", true)
		if err != nil {
			return fmt.Errorf("failed to read password: %v", err)
		}

		err = a.api.Login(username, password)
		if err == nil && a.api.IsLoggedIn() {
			err = a.secrets.SetKey("token", a.api.Token())
			if err != nil {
				return fmt.Errorf("failed to store token: %v", err)
			}

			err = a.secrets.SetKey("userid", a.api.UserId())
			if err != nil {
				return fmt.Errorf("failed to store userid: %v", err)
			}
			err = a.secrets.SetKey("deviceid", a.api.DeviceId)
			if err != nil {
				return fmt.Errorf("failed to store deviceid: %v", err)
			}
			serverId := a.api.ServerId()
			err = a.secrets.SetKey("serverid", serverId)
			if err != nil {
				return fmt.Errorf("failed to store serverid: %v", err)
			}
		} else {
			return fmt.Errorf("login failed")
		}
		return nil

	} else {
		err := a.api.SetToken(token)
		if err != nil {
			return fmt.Errorf("set token: %v", err)
		}
		userid, err := a.secrets.GetKey("userid")
		if err != nil {
			return err
		}
		a.api.SetUserId(userid)

		deviceid, err := a.secrets.GetKey("deviceid")
		if err != nil {
			return err
		}
		a.api.DeviceId = deviceid

		serverId, err := a.secrets.GetKey("serverid")
		if err != nil {
			return err
		}
		a.api.SetServerId(serverId)
		return nil
	}
}

func (a *Application) initApiView() error {
	view, err := a.secrets.GetKey("music_view")
	if err != nil {
		return err
	}
	if view != "" {
		a.api.SetDefaultMusicview(view)
		return nil
	} else {
		views, err := a.api.GetViews()
		if err != nil {
			return fmt.Errorf("get user views: %v", err)
		}
		if len(views) == 0 {
			return fmt.Errorf("no views to use")
		}

		fmt.Println("Found collections: ")
		for i, v := range views {
			fmt.Printf("%d. %s (%s)\n", i+1, v.Name, v.Type)
		}

		// Loop for as long as user gives valid input for default view
		for {
			number, err := config.ReadUserInput("Default music view (number)", false)
			if err != nil {
				fmt.Println("Must be a valid number")
			} else {
				num, err := strconv.Atoi(number)
				if err != nil {
					fmt.Println("Must be a valid number")
				} else {
					id := ""
					if num < len(views) && num > 0 {
						id = views[num].Id.String()
						err = a.secrets.SetKey("music_view", id)
						a.api.SetDefaultMusicview(id)
						if err != nil {
							return err
						}
						return nil
					} else {
						fmt.Println("Must be a valid number")
					}
				}
			}
		}
	}
}

func (a *Application) initApplication() error {
	var err error
	a.player, err = player.NewPlayer(a.api)
	if err != nil {
		return fmt.Errorf("create player: %v", err)
	}
	a.content, err = controller.NewContent(a.api, a.player)
	if err != nil {
		return fmt.Errorf("create content controller: %v", err)
	}

	a.gui = ui.NewUi(a.content)
	return nil
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
