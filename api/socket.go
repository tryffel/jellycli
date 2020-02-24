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
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/url"
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
	return nil
}

func (a *Api) handleSocketInbound(msg interface{}) error {
	logrus.Debug("Received socket message: ", msg)
	return nil

}

func (a *Api) handleSocketOutbount(msg interface{}) error {
	if a.socket == nil {
		return fmt.Errorf("socket not open")
	}
	return a.socket.WriteJSON(msg)
}
