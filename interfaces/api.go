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

package interfaces

import (
	"io"
	"tryffel.net/go/jellycli/models"
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

	Queue []models.Id
}
