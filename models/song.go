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

package models

// Song always belongs to album (even if single) and has artist.
// There might be multiple artists.
type Song struct {
	// Id as unique identifier for song
	Id       Id     `db:"id"`
	Name     string `db:"name"`
	Duration int    `db:"duration"`
	Index    int    `db:"song_index"`
	Album    Id     `db:"album"`
	// DiscNumber tells which disc song is part of
	DiscNumber int `db:"disc_number"`
	// Artists are all artist taking part in song
	Artists []IdName
	// AlbumArtist is primary artist
	AlbumArtist Id `db:"artist"`

	Favorite bool `db:"favorite"`
}

func (s *Song) GetId() Id {
	return s.Id
}

func (s *Song) HasChildren() bool {
	return false
}

func (s *Song) GetChildren() []Id {
	return []Id{}
}

func (s *Song) GetParent() Id {
	return s.Album
}

func (s *Song) GetName() string {
	return s.Name
}

func (s *Song) GetType() ItemType {
	return TypeSong
}

func (s *Song) ToInfo() *SongInfo {
	return &SongInfo{
		Id:       s.Id,
		Name:     s.Name,
		Duration: s.Duration,
		Artist:   "",
		ArtistId: "",
		Album:    "",
		AlbumId:  s.Album,
		Year:     0,
	}

}

type SongInfo struct {
	Id       Id
	Name     string
	Duration int
	Artist   string
	ArtistId Id
	Album    string
	AlbumId  Id
	Year     int
}

func SongsToItems(songs []*Song) []Item {
	if songs == nil {
		return []Item{}
	}
	items := make([]Item, len(songs))

	for i, v := range songs {
		items[i] = v
	}
	return items
}
