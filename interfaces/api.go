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

type Api interface {
	ReportProgress(state *PlaybackState) error
	GetSongDirect(id string, codec string) (io.ReadCloser, error)
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
