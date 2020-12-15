/*
 * Jellycli is a terminal music player for Jellyfin.
 * Copyright (C) 2020 Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
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
	"net/url"
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

	remoteControlEnabled bool
}

func (jf *Jellyfin) AuthOk() error {
	return jf.TokenOk()
}

func (jf *Jellyfin) GetInfo() (*models.ServerInfo, error) {
	info := &models.ServerInfo{
		ServerType: "Jellyfin",
	}

	resp, err := jf.getserverInfo()
	if err != nil {
		return nil, err
	}

	info.Name = resp.ServerName
	info.Id = resp.Id
	info.Version = resp.Version

	if resp.ShutdownPending {
		info.Message = "Shutdown pending"
	} else if resp.RestartPending {
		info.Message = "Restart pending"
	}

	info.Misc = map[string]string{}

	remoteErr := jf.RemoteControlEnabled()
	var remoteStatus string
	if remoteErr != nil {
		remoteStatus = remoteErr.Error()
	} else {
		remoteStatus = "connected"
	}

	info.Misc["Cached objects"] = strconv.Itoa(jf.GetCacheItems())
	info.Misc["Remote control"] = remoteStatus
	return info, nil
}

func (jf *Jellyfin) RemoteControlEnabled() error {
	if !jf.remoteControlEnabled {
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
		jf.host, err = provider.Get("jellyfin.url", false, "jellyfin url")
		if err != nil {
			return jf, err
		}

		if jf.host == "" {
			return jf, errors.New("jellyfin url cannot be empty")
		}

		_, err := url.Parse(jf.host)
		if err != nil {
			return jf, fmt.Errorf("parse url: %v", err)
		}
	}

	err = jf.ping()
	if err != nil {
		logrus.Errorf("connection to jellyfin server failed. Make sure you entered correct url.")
		return jf, fmt.Errorf("connect jellyfin server: %v", err)
	}

	var password string
	if jf.token == "" {
		username, err := provider.Get("jellyfin.username", false, "Username")
		password, err = provider.Get("jellyfin.password", true, "Password")
		if err != nil {
			return jf, err
		}
		err = jf.login(username, password)
		if err != nil {
			return jf, err
		}
	}

	if err = jf.TokenOk(); err != nil {
		if strings.Contains(err.Error(), "invalid token") {
			logrus.Warningf("Authentication required")
			username, err := provider.Get("jellyfin.username", false, "Username")
			password, err = provider.Get("jellyfin.password", true, "Password")
			if err != nil {
				return jf, err
			}
			err = jf.login(username, password)
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
	jf.remoteControlEnabled = true
	jf.player = p
}

func (jf *Jellyfin) SetQueue(q interfaces.QueueController) {
	jf.queue = q
}

func (jf *Jellyfin) ConnectionOk() error {
	info, err := jf.getserverInfo()
	if err != nil {
		return err
	}
	err = jf.VerifyServerId()
	if err != nil {
		return err
	}

	logrus.Debugf("Connected to %s version %s", info.ServerName, info.Version)
	return nil
}

func (jf *Jellyfin) TokenOk() error {
	if jf.token == "" {
		return errors.New("invalid token")
	}
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
		if strings.Contains(err.Error(), "needs authorization") {
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
		number, err := provider.Get("jellyfin.music_view", false, "Default music view (enter number)")
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

func (jf *Jellyfin) ping() error {
	body, err := jf.get("/System/Info/Public", nil)
	if err != nil {
		return err
	}

	type result struct {
		LocalAddress    string
		ServerName      string
		Version         string
		ProductName     string
		OperatingSystem string
		Id              string
	}

	res := &result{}
	err = json.NewDecoder(body).Decode(res)
	if err != nil {
		return fmt.Errorf("invalid json response: %v", err)
	}

	logrus.Debugf("Connect to server %s, (id %s)", res.ServerName, res.Id)
	return nil
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

func (jf *Jellyfin) Start() error {
	err := jf.Connect()
	if err != nil {
		return err
	}
	return jf.Task.Start()
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
