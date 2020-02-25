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

//PlayerState holds data about currently playing song if any
type PlayingState struct {
	State       State
	PlayingType Playtype
	Song        string
	Artist      string
	Album       string
	Year        int

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
	p.Song = ""
	p.Artist = ""
	p.Album = ""
	p.Year = 0
	p.CurrentSongDuration = 0
	p.CurrentSongPast = 0
}

const (
	Play State = iota
	Continue
	SetVolume
	Pause
	Stop
	EndSong
	SongComplete
)

const (
	Song        Playtype     = 0
	Album       Playtype     = 1
	Artist      Playtype     = 2
	Playlist    Playtype     = 3
	StatusOk    PlayerStatus = 0
	StatusError PlayerStatus = 0
)

type Playtype int
type PlayerStatus int

type State int
