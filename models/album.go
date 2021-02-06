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

// Album has multiple songs. It has one primary artist and multiple additional artists.
type Album struct {
	Id       Id     `db:"id"`
	Name     string `db:"name"`
	Year     int    `db:"year"`
	Duration int    `db:"duration"`
	// Artist is the primary artist
	Artist Id `db:"artist"`
	// Additional artists. If length is 1, first item is same as primary artist. Else it contains additional artists
	AdditionalArtists []IdName
	Songs             []Id
	//SongCount, how many songs are there in album.
	// 0 means album is empty, where -1 means songs need to be gathered separately.
	SongCount int `db:"song_count"`
	// ImageId is optional id for image album
	ImageId   string `db:"image_id"`
	DiscCount int    `db:"disc_count"`

	Favorite bool `db:"favorite"`
}

func (a *Album) GetId() Id {
	return a.Id
}

func (a *Album) HasChildren() bool {
	return a.SongCount != 0
}

func (a *Album) GetChildren() []Id {
	return a.Songs
}

func (a *Album) GetParent() Id {
	return a.Artist
}

func (a *Album) GetName() string {
	return a.Name
}

func (a *Album) GetType() ItemType {
	return TypeAlbum
}

func AlbumsToItems(albums []*Album) []Item {
	if albums == nil {
		return []Item{}
	}
	items := make([]Item, len(albums))

	for i, v := range albums {
		items[i] = v
	}
	return items
}
