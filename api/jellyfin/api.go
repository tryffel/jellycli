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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/task"
	"tryffel.net/go/jellycli/util"
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

func (jf *Jellyfin) AuthOk() error {
	return jf.TokenOk()
}

func (jf *Jellyfin) GetInfo() (*models.ServerInfo, error) {
	return &models.ServerInfo{Name: "Jellyfin"}, nil
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

func NewJellyfin(conf *config.Jellyfin, provider config.KeyValueProvider) (*Jellyfin, error) {
	jf := &Jellyfin{
		client: &http.Client{},
	}

	if conf != nil {
		jf.host = conf.Url
		jf.token = conf.Token
		jf.userId = conf.UserId
		jf.serverId = conf.ServerId
		jf.musicView = conf.MusicView
	}

	id, err := machineid.ProtectedID(config.AppName)
	if err != nil {
		return jf, fmt.Errorf("failed to get unique host id: %v", err)
	}
	jf.DeviceId = id
	jf.SessionId = util.RandomKey(15)
	jf.Name = "api"
	jf.SetLoop(jf.loop)

	jf.cache, err = NewCache()
	if err != nil {
		return jf, fmt.Errorf("create cache: %v", err)
	}

	if jf.host == "" {
		jf.host, err = provider.Get("jellyfin host", false, "")
		if err != nil {
			return jf, err
		}
	}

	if jf.token == "" {
		jf.userId, err = provider.Get("username", false, "")
		if err != nil {
			return jf, err
		}
	}

	if err := jf.TokenOk(); err != nil {
		if strings.Contains(err.Error(), "invalid token") {
			logrus.Warningf("Authentication required")
			password, err := provider.Get("Password", true, "")
			if err != nil {
				return jf, err
			}
			err = jf.login(jf.userId, password)
			if err != nil {
				return jf, err
			}
		}
	}

	err = jf.selectDefaultMusicView(provider)
	if err != nil {
		return jf, err
	}

	return jf, err
}

func (jf *Jellyfin) SetPlayer(p interfaces.Player) {
	jf.player = p
}

func (jf *Jellyfin) SetQueue(q interfaces.QueueController) {
	jf.queue = q
}

func (jf *Jellyfin) ConnectionOk() error {
	name, version, _, _, _, err := jf.GetServerVersion()
	if err != nil {
		return err
	}
	err = jf.VerifyServerId()
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

func (jf *Jellyfin) selectDefaultMusicView(provider config.KeyValueProvider) error {
	if jf.musicView != "" {
		return nil
	}
	views, err := jf.GetViews()
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
		number, err := provider.Get("Default music view (enter number)", false, "")
		if err != nil {
			fmt.Println("Must be a valid number")
		} else {
			num, err := strconv.Atoi(number)
			if err != nil {
				fmt.Println("Must be a valid number")
			} else {
				id := ""
				if num < len(views)+1 && num > 0 {
					id = views[num-1].Id.String()
					jf.musicView = id
					if err != nil {
						return err
					}
					return nil
				} else {
					fmt.Println("Must be in range")
				}
			}
		}
	}
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
