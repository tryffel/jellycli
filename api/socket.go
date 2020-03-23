/*
 * Copyright 2020 Tero Vierimaa
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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"math"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

const (
	pongTimeout = 10 * time.Second
	pingPeriod  = (pongTimeout * 9) / 10
)

func (a *Api) connectSocket() error {
	if a.token == "" {
		return fmt.Errorf("no access token")
	}
	u, err := url.Parse(a.host)
	host := u.Host + u.Path
	dialer := websocket.Dialer{
		Proxy:            nil,
		HandshakeTimeout: time.Second * 10,
	}
	logrus.Debug("connecting websocket to ", host)
	socket, _, err := dialer.Dial(
		fmt.Sprintf("wss://%s/socket?api_key=%s&deviceId=%s", host, a.token, a.DeviceId), nil)
	if err != nil {
		a.socketState = socketDisconnected
		return fmt.Errorf("websocket connection failed: %v", err)
	}
	a.socketLock.Lock()
	defer a.socketLock.Unlock()
	logrus.Debugf("websocket connected")
	a.socket = socket

	err = a.socket.SetReadDeadline(time.Now().Add(pongTimeout))
	if err != nil {
		logrus.Errorf("set socket read deadline: %v", err)
	}
	a.socket.SetPongHandler(func(string) error {
		logrus.Trace("Websocket received pong")
		return a.socket.SetReadDeadline(time.Now().Add(pongTimeout))
	})

	a.socketState = socketConnected
	return nil
}

func (a *Api) handleSocketOutbount(msg interface{}) error {
	if a.socket == nil {
		return fmt.Errorf("socket not open")
	}
	return a.socket.WriteJSON(msg)
}

// read next message from socket in blocking mode. Messages are read as long as socket connection is ok
func (a *Api) readMessage() {
	if a.WebsocketOk() {
		msgType, buff, err := a.socket.ReadMessage()
		if err != nil {
			a.handleSocketError(err)
		}
		if msgType == websocket.TextMessage {
			err = a.parseInboudMessage(&buff)
			a.handleSocketError(err)
		}
		go a.readMessage()
	}
}

type webSocketInboudMsg struct {
	MessageType string                 `json:"MessageType"`
	Data        map[string]interface{} `json:"Data"`
}

type controlCommand struct {
	Name      string `json:"Name"`
	Arguments interface{}
}

func (a *Api) parseInboudMessage(buff *[]byte) error {
	msg := webSocketInboudMsg{}
	err := json.Unmarshal(*buff, &msg)
	if err != nil {
		logrus.Errorf("Parse json: %v", err)
		return fmt.Errorf("parse json: %v", err)
	}

	cmd := strings.ToLower(msg.MessageType)
	if cmd == "generalcommand" {
		data := msg.Data
		name := data["Name"]
		ar := data["Arguments"]
		args, ok := ar.(map[string]interface{})
		if ok {
			switch name {
			case "SetVolume":
				vol := args["Volume"]
				volume, err := strconv.Atoi(vol.(string))
				if err != nil {
					logrus.Error("Invalid volume parameter")
				} else {
					volume := interfaces.AudioVolume(volume)
					a.player.SetVolume(volume)
				}
			default:
				logrus.Warning("unknown socket command: ", name)
			}
		} else {
			logrus.Error("unexpected command format from websocket, expected general command args map[string]interface, got", a)
		}
	} else if cmd == "playstate" {
		data := msg.Data
		rawCmd := data["Command"]
		cmd, ok := rawCmd.(string)
		if ok {
			err = a.pushCommand(cmd)
		}
	} else if cmd == "play" {
		var items []string
		if i, ok := msg.Data["ItemIds"].([]interface{}); ok {
			for _, v := range i {
				if id, ok := v.(string); ok {
					items = append(items, id)
				} else {
					logrus.Errorf("remote play, item id is not string: %s", v)
				}
			}
		} else {
			logrus.Error("Received play command, but queue ids are not array. command: ", msg.Data)
		}
		index, ok := msg.Data["StartIndex"].(float64)
		startIndex := 0
		if ok {
			startIndex = int(index)
		}

		command, ok := msg.Data["PlayCommand"].(string)
		if !ok {
			logrus.Error("Received play command, but command is not string: ", msg.Data)
		} else {
			go a.pushSongsToQueue(items[startIndex:], command)
		}
	}
	return err
}

func (a *Api) pushCommand(cmd string) error {
	if a.player == nil {
		return nil
	}

	switch cmd {
	case "PlayPause":
		a.player.PlayPause()
	case "NextTrack":
		a.player.Next()
	case "PreviousTrack":
		a.player.Previous()
	case "Pause":
		a.player.Pause()
	case "Unpause":
		a.player.Continue()
	case "StopMedia":
		a.player.StopMedia()
		a.queue.ClearQueue(true)
	case "Stop":
		a.player.StopMedia()
		a.queue.ClearQueue(true)
	default:
		logrus.Info("Unknown websocket playstate command: ", cmd)
	}
	return nil
}

// handle errors and try reconnecting
func (a *Api) handleSocketError(err error) {
	if err == nil {
		return
	}

	a.socketLock.Lock()
	defer a.socketLock.Unlock()
	if a.socketState == socketReConnecting {
		return
	}

	awaitReconnect := func(reason string) {
		if a.socketState == socketConnected {
			a.socketState = socketAwaitsReconnecting
			logrus.Warning("Websocket closed: ", reason)
		}
	}

	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
		awaitReconnect("going away")
	} else if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
		awaitReconnect("abnormal closure")
	} else if errors.Is(err, syscall.ECONNABORTED) {
		// happens when disconnected from network, e.g. computer on sleep
		awaitReconnect("connection aborted")
	} else if strings.Contains(err.Error(), "i/o timeout") {
		// happens when disconnected from network, e.g. computer on sleep
		awaitReconnect("io timeout")
	} else {
		logrus.Errorf("unknown socket err: %v", err)
	}
}

// WebsocketOk returns true if websocket connection is ok
func (a *Api) WebsocketOk() bool {
	a.socketLock.RLock()
	defer a.socketLock.RUnlock()
	return a.socketState == socketConnected
}

// try reconnecting socket. Return true if success
func (a *Api) reconnectSocket() bool {
	a.socketLock.Lock()
	a.socketState = socketReConnecting
	var err error
	if a.socket != nil {
		err = a.socket.Close()
		if err != nil {
			logrus.Debugf("reconnect socket: close socket: %v", err)
		} else {
			a.socket = nil
		}
	}

	a.socketLock.Unlock()
	err = a.connectSocket()
	if err != nil {
		logrus.Debugf("reconnect socket: %v", err)
		return false
	}
	logrus.Warning("Websocket reconnected")
	return true
}

// push songs to queue.
func (a *Api) pushSongsToQueue(items []string, mode string) {
	ids := []models.Id{}
	for _, v := range items {
		ids = append(ids, models.Id(v))
	}

	var songs []*models.Song
	var err error

	// server does not accept too long id list (> 15 ids), so we need to split large queries
	if len(ids) > 15 {
		rounds := int(math.Ceil(float64(len(ids)) / 15))
		logrus.Infof("Too many songs for single query, split query: %d total, %d queries", len(ids), rounds)
		for i := 0; i < rounds; i++ {
			from := i * 15
			to := (i + 1) * 15
			if to > len(ids) {
				to = len(ids)
			}
			logrus.Debugf("Download songs [%d, %d]", from, to)
			s, err := a.GetSongsById(ids[from:to])
			if err != nil {
				logrus.Errorf("download songs: %v", err)
			}
			songs = append(songs, s...)
		}
		if len(songs) != len(ids) {
			logrus.Errorf("some songs were not downloaded: expect %d, got %d", len(ids), len(songs))
		}
	} else {
		songs, err = a.GetSongsById(ids)
	}

	if err != nil {
		logrus.Errorf("remote control: add songs to queue: get songs from ids: %v", err)
		return
	}
	logrus.Debug("received play event: ", mode)

	// some modes are swapped in other clients, use those for consistency
	if mode == "PlayNow" {
		a.player.StopMedia()
		a.queue.ClearQueue(true)
		a.queue.PlayNext(songs)
	} else if mode == "PlayLast" {
		//} else if mode == "PlayNext" {
		a.queue.PlayNext(songs)
	} else if mode == "PlayNext" {
		//} else if mode == "PlayLast" {
		a.queue.AddSongs(songs)
	} else {
		logrus.Errorf("unknown remote play mode: %s", mode)
	}
}
