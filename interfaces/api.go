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

package interfaces

import (
	"io"
)

type ApiPlaybackEvent string

const (
	// Internal events
	EventStart ApiPlaybackEvent = "start"
	EventStop  ApiPlaybackEvent = "stop"

	// Outgoing events
	EventTimeUpdate          ApiPlaybackEvent = "TimeUpdate"
	EventPause               ApiPlaybackEvent = "Pause"
	EventUnpause             ApiPlaybackEvent = "Unpause"
	EventVolumeChange        ApiPlaybackEvent = "VolumeChange"
	EventRepeatModeChange    ApiPlaybackEvent = "RepeatModeChange"
	EventAudioTrackChange    ApiPlaybackEvent = "AudioTrackChange"
	EventSubtitleTrackChange ApiPlaybackEvent = "SubtitleTrackChange"
	EventPlaylistItemMove    ApiPlaybackEvent = "PlaylistItemMove"
	EventPlaylistItemRemove  ApiPlaybackEvent = "PlaylistItemRemove"
	EventPlaylistItemAdd     ApiPlaybackEvent = "PlaylistItemAdd"
	EventQualityChange       ApiPlaybackEvent = "QualityChange"
)

type Api interface {
	ReportProgress(state *ApiPlaybackState) error
	GetSongDirect(id string, codec string) (io.ReadCloser, error)
}

//Playbackstate reports playback back to server
type ApiPlaybackState struct {
	Event    ApiPlaybackEvent
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
