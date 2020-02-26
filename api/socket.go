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
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	pongTimeout = 60 * time.Second
	pingPeriod  = (pongTimeout * 9) / 10
)

func (a *Api) connectSocket() error {
	if a.token == "" {
		return fmt.Errorf("no access token")
	}
	u, err := url.Parse(a.host)
	host := u.Host + u.Path
	socket, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://%s/socket?api_key=%s&deviceId=%s", host, a.token, a.DeviceId), nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
	}

	logrus.Infof("Websocket ok")
	a.socket = socket

	a.socket.SetReadDeadline(time.Now().Add(pongTimeout))
	a.socket.SetPongHandler(func(string) error {
		logrus.Trace("Websocket received pong")
		return a.socket.SetReadDeadline(time.Now().Add(pongTimeout))
	})
	return nil
}

func (a *Api) handleSocketOutbount(msg interface{}) error {
	if a.socket == nil {
		return fmt.Errorf("socket not open")
	}
	return a.socket.WriteJSON(msg)
}

// read next message from socket in blocking mode
func (a *Api) readMessage() {
	if a.socket != nil {
		msgType, buff, err := a.socket.ReadMessage()
		if err != nil {
			logrus.Errorf("read websocket message: %v", err)
		}
		if msgType == websocket.TextMessage {
			err = a.parseInboudMessage(&buff)
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
					a.controller.SetVolume(volume)
				}
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
	}
	return err
}

func (a *Api) pushCommand(cmd string) error {
	if a.controller == nil {
		return nil
	}

	switch cmd {
	case "PlayPause":
		a.controller.PlayPause()
	case "NextTrack":
		a.controller.Next()
	case "Stop":
		a.controller.StopMedia()
	default:
		logrus.Info("Unknown websocket playstate command: ", cmd)
	}
	return nil
}
