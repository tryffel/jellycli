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

package api

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"tryffel.net/go/jellycli/interfaces"
)

const (
	ticksToSecond = 10000000
)

type infoResponse struct {
	ServerName string `json:"ServerName"`
	Version    string `json:"Version"`
	Id         string `json:"Id"`
}

// GetServerVersion returns name, version, id and possible error
func (a *Api) GetServerVersion() (string, string, string, error) {
	body, err := a.get("/System/Info/Public", nil)
	if err != nil {
		return "", "", "", fmt.Errorf("request failed: %v", err)
	}

	response := infoResponse{}
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return "", "", "", fmt.Errorf("response read failed: %v", err)
	}

	return response.ServerName, response.Version, response.Id, nil
}

func (a *Api) VerifyServerId() error {
	_, _, id, err := a.GetServerVersion()
	if err != nil {
		return err
	}

	if a.serverId != id {
		return fmt.Errorf("server id has changed: expected %s, got %s", a.serverId, id)
	}
	return nil
}

type playbackStarted struct {
	QueueableMediaTypes []string
	CanSeek             bool
	ItemId              string
	MediaSourceId       string
	PositionTicks       int
	VolumeLevel         int
	IsPaused            bool
	IsMuted             bool
	PlayMethod          string
	PlaySessionId       string
	LiveStreamId        string
	PlaylistLength      int
	PlaylistIndex       int
}

type playbackProgress struct {
	playbackStarted
	Event interfaces.ApiPlaybackEvent
}

// ReportProgress reports playback status to server
func (a *Api) ReportProgress(state *interfaces.ApiPlaybackState) error {
	var err error
	var report interface{}
	var url string

	started := playbackStarted{
		QueueableMediaTypes: []string{"Audio"},
		CanSeek:             true,
		ItemId:              state.ItemId,
		MediaSourceId:       state.ItemId,
		PositionTicks:       state.Position * ticksToSecond,
		VolumeLevel:         state.Volume,
		IsPaused:            state.IsPaused,
		IsMuted:             state.IsPaused,
		PlayMethod:          "DirectPlay",
		PlaySessionId:       a.SessionId,
		LiveStreamId:        "",
		PlaylistLength:      state.PlaylistLength * ticksToSecond,
	}

	if state.Event == interfaces.EventStart {
		url = "/Sessions/Playing"
		report = started
	} else if state.Event == interfaces.EventStop {
		url = "/Sessions/Playing/Stopped"
		report = started
	} else {
		url = "/Sessions/Playing/Progress"
		report = playbackProgress{
			playbackStarted: started,
			Event:           state.Event,
		}
	}

	// webui does not accept websocket response for now, so fall back to http posts. No p
	//if a.socket == nil || state.Event == interfaces.EventStart || state.Event == interfaces.EventStop {
	params := *a.defaultParams()
	params["api_key"] = a.token
	body, err := json.Marshal(&report)
	if err != nil {
		return fmt.Errorf("json marshaling failed: %v", err)
	}
	var resp io.ReadCloser
	resp, err = a.post(url, &body, &params)
	resp.Close()

	/*
		} else {
			content := map[string]interface{}{}
			content["MessageType"] = "ReportPlaybackStatus"
			content["Data"] = report

			a.socketLock.Lock()
			a.socket.SetWriteDeadline(time.Now().Add(time.Second * 15))
			err = a.socket.WriteJSON(content)
			a.socketLock.Unlock()
			if err != nil {
				logrus.Errorf("Send playback status via websocket: %v", err)
			}
		}
	*/

	logrus.Debug("Progress event: ", state.Event)

	if err == nil {
		return nil
	} else {
		return fmt.Errorf("push progress: %v", err)
	}
}

func (a *Api) GetCacheItems() int {
	return a.cache.Count()
}

//ImageUrl returns primary image url for item, if there is one. Otherwise return empty
func (a *Api) ImageUrl(item, imageTag string) string {
	return fmt.Sprintf("%s/Items/%s/Images/Primary?maxHeight=500&tag=%s&quality=90", a.host, item, imageTag)
}

func (a *Api) ReportCapabilities() error {
	data := map[string]interface{}{}
	data["PlayableMediaTypes"] = []string{"Audio"}
	data["SupportedCommands"] = []string{
		"VolumeUp",
		"VolumeDown",
		"Mute",
		"Unmute",
		"ToggleMute",
		"SetVolume",
	}
	data["SupportsMediaControl"] = true
	data["SupportsPersistentIdentifier"] = false

	params := *a.defaultParams()
	params["api_key"] = a.token

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json: %v", err)
	}

	url := "/Sessions/Capabilities/Full"
	resp, err := a.post(url, &body, &params)
	if err != nil {
		return err
	}
	resp.Close()
	return nil
}
