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
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
)

// Subsonic implements subsonic api.
type Subsonic struct {
	host       string
	salt       string
	token      string
	user       string
	apiversion string
	client     string

	connectionStatus string
	connectionError  *subError

	musicFolder int

	favoriteArtists []*models.Artist
	favoriteAlbums  []*models.Album
}

func (s *Subsonic) Stream(Song *models.Song) (io.ReadCloser, interfaces.AudioFormat, error) {
	params := &params{}
	params.setId(Song.Id.String())
	(*params)["estimateContentLength"] = "true"
	(*params)["s"] = s.salt
	(*params)["t"] = s.token
	(*params)["u"] = s.user
	(*params)["c"] = s.client
	(*params)["v"] = s.apiversion

	url := s.host + "/rest/stream"

	stream, err := api.NewStreamDownload(url, nil, *params, http.DefaultClient, Song.Duration)
	if err != nil {
		return nil, interfaces.AudioFormatNil, err
	}

	format, err := stream.AudioFormat()
	return stream, format, err
}

func (s *Subsonic) Download(Song *models.Song) (io.ReadCloser, interfaces.AudioFormat, error) {
	return s.Stream(Song)
}

func (s *Subsonic) GetInfo() (*models.ServerInfo, error) {
	info := &models.ServerInfo{
		ServerType: "Subsonic",
	}

	resp, err := s.get("/ping", nil)
	if err != nil {
		return nil, err
	}

	info.Id = resp.Type
	info.Name = resp.Type
	info.Version = resp.ServerVersion
	return info, nil
}

func (s *Subsonic) ConnectionOk() error {
	if s.connectionError != nil {
		return fmt.Errorf("subsonic error: (%d): %s", s.connectionError.Code, s.connectionError.Message)
	}
	return nil
}

func NewSubsonic(conf *config.Subsonic, provider config.KeyValueProvider) (*Subsonic, error) {
	s := &Subsonic{
		host:       conf.Url,
		salt:       conf.Salt,
		token:      conf.Token,
		user:       conf.Username,
		apiversion: "1.16.1",
		client:     "Jellycli",
	}

	if s.host == "" {
		host, err := provider.Get("subsonic host", false, "")
		if err != nil {
			return s, err
		}
		if host != "" {
			s.host = host
		} else {
			return s, errors.New("subsonic host cannot be empty")
		}
	}

	err := s.checkConnection()
	if err != nil {
		loginErr := s.login(provider)
		if loginErr != nil {

			return s, loginErr
		}

		connErr := s.checkConnection()
		if connErr != nil {
			return s, err
		}
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
	s.connectionError = nil
	return s, nil
}

func (s *Subsonic) get(url string, params *params) (*response, error) {
	fullUrl := s.host + "/rest" + url
	start := time.Now()
	req, _ := http.NewRequest(http.MethodGet, fullUrl, nil)

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
	took := time.Now().Sub(start)
	if err != nil {
		logrus.Warningf("Get %s failed", "/rest"+url)
		return nil, err
	}

	logrus.Debugf("Get %s status: %d, took: %d ms", "/rest"+url, resp.StatusCode, took.Milliseconds())
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
		if resp != nil {
			s.connectionError = resp.Error
		}
		return err
	}

	if resp.Status == "ok" {
		return nil
	}
	return fmt.Errorf("invalid server status: %s, expected 'ok'", resp.Status)
}

func (s *Subsonic) login(provider config.KeyValueProvider) error {

	logrus.Warning("Authentication required for Subsonic")

	username, err := provider.Get("username", false, "")
	if err != nil {
		return err
	}
	password, err := provider.Get("password", true, "")
	if err != nil {
		return err
	}

	s.user = username
	s.salt = util.RandomKey(15)
	s.token = fmt.Sprintf("%x", md5.Sum([]byte(password+s.salt)))
	return nil
}

type params map[string]string

func (p *params) setId(id string) {
	(*p)["id"] = id
}

func (s *Subsonic) GetConfig() config.Backend {
	return &config.Subsonic{
		Url:      s.host,
		Username: s.user,
		Salt:     s.salt,
		Token:    s.token,
	}
}

func (s *Subsonic) Start() error {
	return nil
}

func (s *Subsonic) Stop() error {
	return nil
}
