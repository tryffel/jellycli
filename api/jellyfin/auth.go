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
	if err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
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
