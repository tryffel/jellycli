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

import "tryffel.net/go/jellycli/models"

//PlayerState holds data about currently playing song if any
type PlayingState struct {
	State         State
	PlayingType   Playtype
	Song          *models.Song
	Album         *models.Album
	Artist        *models.Artist
	AlbumImageUrl string

	// Content duration in sec
	CurrentSongDuration int
	CurrentSongPast     int
	PlaylistDuration    int
	PlaylistLeft        int
	// Volume [0,100]
	Volume int
}

//remove song info from state
func (p *PlayingState) Clear() {
	p.Song = nil
	p.Album = nil
	p.Artist = nil
	p.AlbumImageUrl = ""
	p.CurrentSongDuration = 0
	p.CurrentSongPast = 0
}

const (
	// Player states
	// StopMedia -> Play -> Pause -> (Continue) -> StopMedia
	// Play new song
	Play State = iota
	// Continue paused song, only a transition mode, never state of the player
	Continue
	//SetVolume, only transition mode
	SetVolume
	// Pause song
	Pause
	// StopMedia playing
	Stop
	//EndSong is a transition state to end current song
	EndSong
	//SongComplete, only transition mode to notify song has changed
	SongComplete
)

const (
	// Playing single song
	Song Playtype = 0
	// Playing album
	Album Playtype = 1
	// Playing artists discography
	Artist Playtype = 2
	// Playing playlist
	Playlist Playtype = 3
	// Last action was ok
	StatusOk PlayerStatus = 0
	// Last action resulted in error
	StatusError PlayerStatus = 0
)

type Playtype int
type PlayerStatus int

type State int
