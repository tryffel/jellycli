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
	"os"
	"tryffel.net/pkg/jellycli/api"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/ui"
)

func main() {
	conf, err := config.NewSecretStore()
	if err != nil {
		fmt.Println("Failed to start application:", err)
		os.Exit(1)
	}

	host, err := conf.EnsureKey("jellyfin_host")
	if err != nil {
		fmt.Printf("Failed to get jellyfin host: %v", err)
	}

	fmt.Println("Connecting to ", host)
	client := api.NewApi(host)

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
			fmt.Println(fmt.Errorf("failed to set wallet value: %v", err))
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Printf("failure in login: %v", err)
	}

	client.GetUserViews()

	gui, err := ui.NewGui()
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	err = gui.Show()
	if err != nil && err != gocui.ErrQuit {
		fmt.Printf("Gui error: %v", err)
	}
}
