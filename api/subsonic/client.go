/*
 * Copyright 2020 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package subsonic contains remote server implementation for Subsonic-compatible servers.
// Implemented: api.Browser.
// Subsonic-protocol does not support api.RemoteController.
package subsonic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"tryffel.net/go/jellycli/models"
)

// Subsonic implements subsonic api.
type Subsonic struct {
	host       string
	salt       string
	token      string
	user       string
	apiversion string
	client     string

	musicFolder int

	favoriteArtists []*models.Artist
	favoriteAlbums  []*models.Album
}

func NewSubsonic(url, user, salt, token string) (*Subsonic, error) {
	s := &Subsonic{
		host:       url,
		salt:       salt,
		token:      token,
		user:       user,
		apiversion: "1.16.1",
		client:     "Jellycli",
	}
	err := s.checkConnection()
	if err != nil {
		return s, err
	}

	resp, err := s.get("/getMusicFolders", nil)
	if err != nil {
		return s, fmt.Errorf("get music folders: %v", err)
	}

	if resp.MusicFolders == nil {
		return s, errors.New("no music folders available")
	}

	if len(resp.MusicFolders.Folders) == 0 {
		return s, errors.New("no music folders available")
	}

	s.musicFolder = resp.MusicFolders.Folders[0].Id
	return s, nil
}

func (s *Subsonic) get(url string, params *params) (*response, error) {
	req, _ := http.NewRequest(http.MethodGet, s.host+url, nil)

	q := req.URL.Query()
	q.Add("s", s.salt)
	q.Add("t", s.token)
	q.Add("u", s.user)
	q.Add("c", s.client)
	q.Add("v", s.apiversion)
	q.Add("f", "json")

	if params != nil {
		for key, value := range *params {
			q.Add(key, value)
		}
	}

	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dto := &subResponse{}
	err = json.NewDecoder(resp.Body).Decode(dto)

	if err != nil {
		return dto.Resp, err
	}

	if dto.Resp.Status != "ok" {
		err = fmt.Errorf("subsonic: (%d - %s) %s", dto.Resp.Error.Code, dto.Resp.Error.Code.String(),
			dto.Resp.Error.Message)
	}
	return dto.Resp, err
}

func (s *Subsonic) checkConnection() error {
	resp, err := s.get("/ping", nil)
	if err != nil {
		return err
	}

	if resp.Status == "ok" {
		return nil
	}
	return fmt.Errorf("invalid server status: %s, expected 'ok'", resp.Status)
}

type params map[string]string

func (p *params) setId(id string) {
	(*p)["id"] = id
}
