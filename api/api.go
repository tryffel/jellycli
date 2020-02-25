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

package api

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/task"
)

type Api struct {
	task.Task
	cache     *Cache
	host      string
	token     string
	userId    string
	serverId  string
	DeviceId  string
	SessionId string
	client    *http.Client
	loggedIn  bool
	musicView string

	controller interfaces.MediaController

	socket     *websocket.Conn
	socketChan chan interface{}
}

func NewApi(host string) (*Api, error) {
	a := &Api{
		host:       host,
		token:      "",
		client:     &http.Client{},
		socketChan: make(chan interface{}),
	}

	id, err := machineid.ProtectedID(config.AppName)
	if err != nil {
		return a, fmt.Errorf("failed to get unique host id: %v", err)
	}
	a.DeviceId = id
	a.SessionId = randomKey(15)
	a.Name = "api"
	a.SetLoop(a.loop)

	a.cache, err = NewCache()
	if err != nil {
		return a, fmt.Errorf("create cache: %v", err)
	}

	return a, nil
}

func (a *Api) SetController(c interfaces.MediaController) {
	a.controller = c
}

func (a *Api) Host() string {
	return a.host
}

func (a *Api) Token() string {
	return a.token
}

//Login performs username based login
func (a *Api) Login(username, password string) error {
	return a.login(username, password)
}

//SetToken sets existing token
func (a *Api) SetToken(token string) error {
	a.token = token
	return nil
}

func (a *Api) tokenExists() error {
	if a.token == "" {
		return errors.New("not logged in")
	}
	return nil
}

func (a *Api) SetUserId(id string) {
	a.userId = id
}

func (a *Api) UserId() string {
	return a.userId
}

func (a *Api) IsLoggedIn() bool {
	return a.loggedIn
}

func (a *Api) ConnectionOk() bool {
	name, version, _, err := a.GetServerVersion()
	if err != nil {
		logrus.Errorf("No connection to server: %v", err)
		return false
	}

	logrus.Infof("Connected to %s version %s", name, version)
	return true
}

func (a *Api) DefaultMusicView() string {
	return a.musicView
}

func (a *Api) SetDefaultMusicview(id string) {
	a.musicView = id
}

func (a *Api) ServerId() string {
	return a.serverId
}

func (a *Api) SetServerId(id string) {
	a.serverId = id
}

// Connect opens a connection to server. If websockets are supported, use that. Report capabilities to server.
// This should be called before streaming any media
func (a *Api) Connect() error {

	var err error
	err = a.ReportCapabilities()
	if err != nil {
		logrus.Warningf("report capabilities: %v", err)
	}

	err = a.connectSocket()
	if err != nil {
		logrus.Infof("No websocket connection: %v", err)
	}

	return nil
}

func (a *Api) loop() {
	if a.socket == nil {
		return
	}

	pingTicker := time.NewTicker(pingPeriod)

	go a.readMessage()
	for true {
		select {
		case <-a.StopChan():
			break
		case _ = <-a.socketChan:
		case <-pingTicker.C:
			logrus.Tracef("Websocket send ping")
			timeout := time.Now().Add(time.Second * 15)
			a.socket.SetWriteDeadline(timeout)
			a.socket.WriteControl(websocket.PingMessage, []byte{}, timeout)
		}
	}

	err := a.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		logrus.Errorf("close websocket: %v", err)
	}
}

func getBodyMsg(body io.ReadCloser) string {
	if body == nil {
		return ""
	}

	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return ""
	}

	return string(bytes)
}

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

func randomKey(length int) string {
	r := rand.Reader
	data := make([]byte, length)
	r.Read(data)

	for i, b := range data {
		data[i] = letters[b%byte(len(letters))]
	}
	return string(data)
}
