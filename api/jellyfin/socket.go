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

package jellyfin

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

func (jf *Jellyfin) connectSocket() error {
	if jf.token == "" {
		return fmt.Errorf("no access token")
	}
	u, err := url.Parse(jf.host)
	host := u.Host + u.Path
	scheme := "wss"
	if u.Scheme == "http" {
		logrus.Info("Websocket encryption disabled")
		scheme = "ws"
	}
	dialer := websocket.Dialer{
		Proxy:            nil,
		HandshakeTimeout: time.Second * 10,
	}
	logrus.Debug("connecting websocket to ", host)
	socket, _, err := dialer.Dial(
		fmt.Sprintf("%s://%s/socket?api_key=%s&deviceId=%s", scheme, host, jf.token, jf.DeviceId), nil)
	if err != nil {
		jf.socketState = socketDisconnected
		return fmt.Errorf("websocket connection failed: %v", err)
	}
	jf.socketLock.Lock()
	defer jf.socketLock.Unlock()
	logrus.Debugf("websocket connected")
	jf.socket = socket

	err = jf.socket.SetReadDeadline(time.Now().Add(pongTimeout))
	if err != nil {
		logrus.Errorf("set socket read deadline: %v", err)
	}
	jf.socket.SetPongHandler(func(string) error {
		logrus.Trace("Websocket received pong")
		return jf.socket.SetReadDeadline(time.Now().Add(pongTimeout))
	})

	jf.socketState = socketConnected
	return nil
}

func (jf *Jellyfin) handleSocketOutbount(msg interface{}) error {
	if jf.socket == nil {
		return fmt.Errorf("socket not open")
	}
	return jf.socket.WriteJSON(msg)
}

// read next message from socket in blocking mode. Messages are read as long as socket connection is ok
func (jf *Jellyfin) readMessage() {
	if jf.WebsocketOk() {
		msgType, buff, err := jf.socket.ReadMessage()
		if err != nil {
			jf.handleSocketError(err)
		}
		if msgType == websocket.TextMessage {
			err = jf.parseInboudMessage(&buff)
			jf.handleSocketError(err)
		}
		go jf.readMessage()
	}
}

type webSocketInboudMsg struct {
	MessageType string      `json:"MessageType"`
	Data        interface{} `json:"Data"`
}

type controlCommand struct {
	Name      string `json:"Name"`
	Arguments interface{}
}

func (jf *Jellyfin) parseInboudMessage(buff *[]byte) error {
	msg := webSocketInboudMsg{}
	err := json.Unmarshal(*buff, &msg)
	if err != nil {
		logrus.Errorf("Parse json: %v", err)

		str := string(*buff)
		logrus.Error(str)
		return fmt.Errorf("parse json: %v, body: %s", err, str)
	}

	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		if msg.MessageType != "ForceKeepAlive" {
			logrus.Errorf("Unknown websocket event: %v", msg)
		}
		return nil
	}

	cmd := strings.ToLower(msg.MessageType)
	if cmd == "generalcommand" {
		name := dataMap["Name"]
		ar := dataMap["Arguments"]
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
					jf.player.SetVolume(volume)
				}
			case "ToggleMute":
				jf.player.ToggleMute()
			default:
				logrus.Warning("unknown socket command: ", name)
			}
		} else {
			logrus.Error("unexpected command format from websocket, expected general command args map[string]interface, got", jf)
		}
	} else if cmd == "playstate" {
		rawCmd := dataMap["Command"]
		cmd, ok := rawCmd.(string)
		if ok {
			err = jf.pushCommand(cmd)
		}
	} else if cmd == "play" {
		var items []string
		if i, ok := dataMap["ItemIds"].([]interface{}); ok {
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
		index, ok := dataMap["StartIndex"].(float64)
		startIndex := 0
		if ok {
			startIndex = int(index)
		}

		command, ok := dataMap["PlayCommand"].(string)
		if !ok {
			logrus.Error("Received play command, but command is not string: ", msg.Data)
		} else {
			go jf.pushSongsToQueue(items[startIndex:], command)
		}
	}
	return err
}

func (jf *Jellyfin) pushCommand(cmd string) error {
	if jf.player == nil {
		return nil
	}

	switch cmd {
	case "PlayPause":
		jf.player.PlayPause()
	case "NextTrack":
		jf.player.Next()
	case "PreviousTrack":
		jf.player.Previous()
	case "Pause":
		jf.player.Pause()
	case "Unpause":
		jf.player.Continue()
	case "StopMedia":
		jf.player.StopMedia()
		jf.queue.ClearQueue(true)
	case "Stop":
		jf.player.StopMedia()
		jf.queue.ClearQueue(true)
	default:
		logrus.Info("Unknown websocket playstate command: ", cmd)
	}
	return nil
}

// handle errors and try reconnecting
func (jf *Jellyfin) handleSocketError(err error) {
	if err == nil {
		return
	}

	jf.socketLock.Lock()
	defer jf.socketLock.Unlock()
	if jf.socketState == socketReConnecting {
		return
	}

	awaitReconnect := func(reason string) {
		if jf.socketState == socketConnected {
			jf.socketState = socketAwaitsReconnecting
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
func (jf *Jellyfin) WebsocketOk() bool {
	jf.socketLock.RLock()
	defer jf.socketLock.RUnlock()
	return jf.socketState == socketConnected
}

// try reconnecting socket. Return true if success
func (jf *Jellyfin) reconnectSocket() bool {
	jf.socketLock.Lock()
	jf.socketState = socketReConnecting
	var err error
	if jf.socket != nil {
		err = jf.socket.Close()
		if err != nil {
			logrus.Debugf("reconnect socket: close socket: %v", err)
		} else {
			jf.socket = nil
		}
	}

	jf.socketLock.Unlock()
	err = jf.connectSocket()
	if err != nil {
		logrus.Debugf("reconnect socket: %v", err)
		return false
	}
	logrus.Warning("Websocket reconnected")
	return true
}

// push songs to queue.
func (jf *Jellyfin) pushSongsToQueue(items []string, mode string) {
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
			s, err := jf.GetSongsById(ids[from:to])
			if err != nil {
				logrus.Errorf("download songs: %v", err)
			}
			songs = append(songs, s...)
		}
		if len(songs) != len(ids) {
			logrus.Errorf("some songs were not downloaded: expect %d, got %d", len(ids), len(songs))
		}
	} else {
		songs, err = jf.GetSongsById(ids)
	}

	if err != nil {
		logrus.Errorf("remote control: add songs to queue: get songs from ids: %v", err)
		return
	}
	logrus.Debug("received play event: ", mode)

	// some modes are swapped in other clients, use those for consistency
	if mode == "PlayNow" {
		jf.player.StopMedia()
		jf.queue.ClearQueue(true)
		jf.queue.PlayNext(songs)
	} else if mode == "PlayLast" {
		//} else if mode == "PlayNext" {
		jf.queue.PlayNext(songs)
	} else if mode == "PlayNext" {
		//} else if mode == "PlayLast" {
		jf.queue.AddSongs(songs)
	} else {
		logrus.Errorf("unknown remote play mode: %s", mode)
	}
}
