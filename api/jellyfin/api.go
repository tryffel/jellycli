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

// Package jellyfin implements connection to Jellyfin server.
// It supports websocket for receiving commands from server and updating status via http post.
package jellyfin

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/task"
)

type socketState int

const (
	socketConnected socketState = iota
	socketDisconnected
	socketReConnecting
	socketAwaitsReconnecting
)

type Jellyfin struct {
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

	player interfaces.Player
	queue  interfaces.QueueController

	socketLock  sync.RWMutex
	socket      *websocket.Conn
	socketState socketState

	enableRemoteControl bool
}

func (jf *Jellyfin) GetServerInfo() api.ServerInfo {
	return api.ServerInfo{Name: "Jellyfin"}
}

func (jf *Jellyfin) RemoteControlEnabled() error {
	if !jf.enableRemoteControl {
		return errors.New("disabled by user")
	}

	jf.socketLock.RLock()
	defer jf.socketLock.RUnlock()

	switch jf.socketState {
	case socketAwaitsReconnecting, socketReConnecting:
		return errors.New("connecting")
	case socketConnected:
		return nil
	case socketDisconnected:
		return errors.New("unable to connect")
	}

	return errors.New("failure")
}

func NewApi(host string, allowRemoteControl bool) (*Jellyfin, error) {
	a := &Jellyfin{
		host:                host,
		token:               "",
		client:              &http.Client{},
		enableRemoteControl: allowRemoteControl,
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

func (jf *Jellyfin) SetPlayer(p interfaces.Player) {
	jf.player = p
}

func (jf *Jellyfin) SetQueue(q interfaces.QueueController) {
	jf.queue = q
}

func (jf *Jellyfin) Host() string {
	return jf.host
}

func (jf *Jellyfin) Token() string {
	return jf.token
}

//Login performs username based login
func (jf *Jellyfin) Login(username, password string) error {
	return jf.login(username, password)
}

//SetToken sets existing token
func (jf *Jellyfin) SetToken(token string) error {
	jf.token = token
	return jf.TokenOk()
}

func (jf *Jellyfin) tokenExists() error {
	if jf.token == "" {
		return errors.New("not logged in")
	}
	return nil
}

func (jf *Jellyfin) SetUserId(id string) {
	jf.userId = id
}

func (jf *Jellyfin) UserId() string {
	return jf.userId
}

func (jf *Jellyfin) IsLoggedIn() bool {
	return jf.loggedIn
}

func (jf *Jellyfin) ConnectionOk() error {
	name, version, _, _, _, err := jf.GetServerVersion()
	if err != nil {
		return err
	}

	logrus.Debugf("Connected to %s version %s", name, version)
	return nil
}

func (jf *Jellyfin) TokenOk() error {
	type serverInfo struct {
		SystemUpdateLevel string `json:"SystemUpdateLevel"`
		RestartPending    bool   `json:"HasPendingRestart"`
		IsShuttingDown    bool   `json:"IsShuttingDown"`
	}

	// check token validity
	body, err := jf.get("/System/Info", nil)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		if strings.Contains(err.Error(), "Access token is invalid or expired") {
			return fmt.Errorf("invalid token: %v", err)
		}
		return err
	}

	info := serverInfo{}
	err = json.NewDecoder(body).Decode(&info)
	if err != nil {
		return fmt.Errorf("decode json: %v", err)
	}
	return nil
}

func (jf *Jellyfin) DefaultMusicView() string {
	return jf.musicView
}

func (jf *Jellyfin) SetDefaultMusicview(id string) {
	jf.musicView = id
}

func (jf *Jellyfin) ServerId() string {
	return jf.serverId
}

func (jf *Jellyfin) SetServerId(id string) {
	jf.serverId = id
}

// Connect opens a connection to server. If websockets are supported, use that. Report capabilities to server.
// This should be called before streaming any media
func (jf *Jellyfin) Connect() error {
	err := jf.ReportCapabilities()
	if err != nil {
		return fmt.Errorf("report capabilities: %v", err)
	}
	err = jf.connectSocket()
	if err != nil {
		logrus.Infof("No websocket connection: %v", err)
	}

	return nil
}

func (jf *Jellyfin) loop() {
	if jf.socket == nil {
		return
	}

	pingTicker := time.NewTicker(pingPeriod)

	// how often to check socket state
	socketTimer := time.NewTimer(time.Second * 2)
	// backoff for reconnecting socket
	socketBackOff := time.Second

	go jf.readMessage()
	for true {
		select {
		case <-jf.StopChan():
			break
		case <-pingTicker.C:
			logrus.Tracef("Websocket send ping")
			timeout := time.Now().Add(time.Second * 15)
			jf.socketLock.Lock()
			if jf.socketState == socketConnected {
				err := jf.socket.SetWriteDeadline(timeout)
				if err != nil {
					logrus.Errorf("set socket write deadline: %v", err)
					jf.handleSocketError(err)
				}
				err = jf.socket.WriteControl(websocket.PingMessage, []byte{}, timeout)
				if err != nil {
					logrus.Errorf("send ping to socket: %v", err)
					jf.handleSocketError(err)
				}
			}
			jf.socketLock.Unlock()
		// keep websocket connected if possible
		case <-socketTimer.C:
			jf.socketLock.RLock()
			state := jf.socketState
			jf.socketLock.RUnlock()
			if state == socketConnected {
				// no worries
				socketTimer.Reset(time.Second * 2)
				socketBackOff = time.Second * 2
			} else if state == socketReConnecting {
				// await
				socketTimer.Reset(time.Second)
				logrus.Debug("websocket reconnect ongoing")
			} else if state == socketAwaitsReconnecting || state == socketDisconnected {
				// start reconnection

				ok := jf.reconnectSocket()
				if ok {
					socketTimer.Reset(time.Second)
					socketBackOff = 0
					go jf.readMessage()
				} else {
					socketBackOff *= 2
					if socketBackOff > time.Second*30 {
						socketBackOff = time.Second * 30
					}
					socketTimer.Reset(socketBackOff)
					logrus.Debugf("websocket reconnection failed, retry after %s", socketBackOff.String())
				}
			}
		}
	}

	err := jf.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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
