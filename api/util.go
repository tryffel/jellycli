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

type PlaybackEvent string

const (
	// Internal events
	EventStart PlaybackEvent = "start"
	EventStop  PlaybackEvent = "stop"

	// Outgoing events
	EventTimeUpdate          PlaybackEvent = "TimeUpdate"
	EventPause               PlaybackEvent = "Pause"
	EventUnpause             PlaybackEvent = "Unnpause"
	EventVolumeChange        PlaybackEvent = "VolumeChange"
	EventRepeatModeChange    PlaybackEvent = "RepeatModeChange"
	EventAudioTrackChange    PlaybackEvent = "AudioTrackChange"
	EventSubtitleTrackChange PlaybackEvent = "SubtitleTrackChange"
	EventPlaylistItemMove    PlaybackEvent = "PlaylistItemMove"
	EventPlaylistItemRemove  PlaybackEvent = "PlaylistItemRemove"
	EventPlaylistItemAdd     PlaybackEvent = "PlaylistItemAdd"
	EventQualityChange       PlaybackEvent = "QualityChange"
)

type playbackProgress struct {
	playbackStarted
	Event PlaybackEvent
}

//Playbackstate reports playback back to server
type PlaybackState struct {
	Event    PlaybackEvent
	ItemId   string
	IsPaused bool
	IsMuted  bool
	// Total length of current playlist in seconds
	PlaylistLength int
	// Position in seconds
	Position int
	// Volume in 0-100
	Volume int
}

func (a *Api) ReportProgress(state *PlaybackState) error {
	params := *a.defaultParams()
	params["api_key"] = a.token
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

	if state.Event == EventStart {
		url = "/Sessions/Playing"
		report = started
	} else if state.Event == EventStop {
		url = "/Sessions/Playing/Stopped"
		report = started
	} else {
		url = "/Sessions/Playing/Progress"
		report = playbackProgress{
			playbackStarted: started,
			Event:           state.Event,
		}
	}

	logrus.Debug("Progress event: ", state.Event)

	body, err := json.Marshal(&report)

	//logrus.Debug(string(body))
	if err != nil {
		return fmt.Errorf("json marshaling failed: %v", err)
	}

	_, err = a.post(url, body, &params)
	if err == nil {
		return nil
	} else {
		return fmt.Errorf("failed to post progress: %v", err)
	}
}

func (a *Api) GetCacheItems() int {
	return a.cache.Count()
}
