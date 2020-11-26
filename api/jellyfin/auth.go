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

package jellyfin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"tryffel.net/go/jellycli/config"
)

type loginResponse struct {
	User     userResponse `json:"User"`
	Token    string       `json:"AccessToken"`
	ServerId string       `json:"ServerId"`
}

type userResponse struct {
	Name     string `json:"Name"`
	ServerId string `json:"ServerId"`
	UserId   string `json:"Id"`
}

func (jf *Jellyfin) login(username, password string) error {
	body := map[string]string{}
	body["Username"] = username
	body["PW"] = password

	b := &bytes.Buffer{}
	err := json.NewEncoder(b).Encode(body)

	auth := jf.authHeader()
	req, err := http.NewRequest("POST", jf.host+"/Users/authenticatebyname", b)
	if err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	req.Header.Set("X-Emby-Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := jf.client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		jf.loggedIn = true
		dto := loginResponse{}
		err := json.NewDecoder(resp.Body).Decode(&dto)
		if err != nil {
			return fmt.Errorf("invalid login response: %v", err)
		}

		jf.token = dto.Token
		jf.serverId = dto.ServerId
		jf.userId = dto.User.UserId
		jf.loggedIn = true
		break
	case http.StatusBadRequest:
		reason, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("login failed: %v", err)
		} else {
			return fmt.Errorf("login failed: %s", reason)
		}
	default:
		reason, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("login failed: %v", err)
		} else {
			return fmt.Errorf("login failed: %s", reason)
		}
	}
	return nil
}

func (jf *Jellyfin) GetConfig() config.Backend {
	return &config.Jellyfin{
		Url:       jf.host,
		Token:     jf.token,
		UserId:    jf.userId,
		DeviceId:  jf.DeviceId,
		ServerId:  jf.ServerId(),
		MusicView: jf.musicView,
	}
}
