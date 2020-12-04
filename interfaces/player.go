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

import "tryffel.net/go/jellycli/models"

// AudioState is audio player state, playing song, stopped
type AudioState int

const (
	// AudioStateStopped, no audio to play
	AudioStateStopped AudioState = iota
	// AudioStatePlaying, playing song
	AudioStatePlaying
)

// AudioAction is an action for audio player, set volume, go to next
type AudioAction int

const (
	// AudioActionTimeUpdate means timed update and no actual action has been taken
	AudioActionTimeUpdate AudioAction = iota
	// AudioActionStop stops playing or paused player
	AudioActionStop
	// AudioActionPlay starts stopped player
	AudioActionPlay
	// AudioActionPlayPause toggles play/pause
	AudioActionPlayPause
	// AudioActionNext plays next song from queue
	AudioActionNext
	// AudioActionPrevious plays previous song from queue
	AudioActionPrevious
	// AudioActionSeek seeks song
	AudioActionSeek
	// AudioActionSetVolume sets volume
	AudioActionSetVolume
)

// AudioTick is alias for millisecond
type AudioTick int

func (a AudioTick) Seconds() int {
	return int(a / 1000)
}

func (a AudioTick) MilliSeconds() int {
	return int(a)
}

func (a AudioTick) MicroSeconds() int {
	return int(a) * 1000
}

// AudioVolume is volume level in [0,100]
type AudioVolume int

const (
	AudioVolumeMax = 100
	AudioVolumeMin = 0
)

// InRange returns true if volume is in allowed range
func (a AudioVolume) InRange() bool {
	return a >= AudioVolumeMin && a <= AudioVolumeMax
}

// Add adds value to volume. Negative values are allowed. Always returns volume that's in allowed range.
func (a AudioVolume) Add(vol int) AudioVolume {
	result := a + AudioVolume(vol)
	if result < AudioVolumeMin {
		return AudioVolumeMin
	}
	if result > AudioVolumeMax {
		return AudioVolumeMax
	}
	return result
}

// AudioStatus contains audio player status
type AudioStatus struct {
	State  AudioState
	Action AudioAction

	Song          *models.Song
	Album         *models.Album
	Artist        *models.Artist
	AlbumImageUrl string

	SongPast AudioTick
	Volume   AudioVolume
	Muted    bool
	Paused   bool
}

func (a *AudioStatus) Clear() {
	a.Song = nil
	a.Album = nil
	a.Artist = nil
	a.AlbumImageUrl = ""
	a.SongPast = 0
	a.Volume = 0
}

// Player controls media playback. Current status is sent to StatusCallback, if set. Multiple status callbacks
// can be set.
type Player interface {
	//PlayPause toggles pause
	PlayPause()
	//Pause pauses media that's currently playing. If none, do nothing.
	Pause()
	//Continue continues currently paused media.
	Continue()
	//StopMedia stops playing media.
	StopMedia()
	//Next plays currently next item in queue. If there's no next song available, this method does nothing.
	Next()
	//Previous plays last played song (first in history) if there is one.
	Previous()
	//Seek seeks forward given seconds
	Seek(ticks AudioTick)
	//SeekBackwards seeks backwards given seconds
	//AddStatusCallback adds callback that get's called every time status has changed,
	//including playback progress
	AddStatusCallback(func(status AudioStatus))
	//SetVolume sets volume to given level in range of [0,100]
	SetVolume(volume AudioVolume)
	// SetMute mutes or un-mutes audio
	SetMute(muted bool)
}
