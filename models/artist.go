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

// Artist has multiple albums.
type Artist struct {
	Id            Id     `db:"id"`
	Name          string `db:"name"`
	Albums        []Id
	TotalDuration int `db:"total_duration"`
	AlbumCount    int `db:"album_count"`

	Favorite bool `db:"favorite"`
}

func (a *Artist) GetId() Id {
	return a.Id
}

func (a *Artist) HasChildren() bool {
	return a.AlbumCount != 0
}

func (a *Artist) GetChildren() []Id {
	return a.Albums
}

func (a *Artist) GetParent() Id {
	return ""
}

func (a *Artist) GetName() string {
	return a.Name
}

func (a *Artist) GetType() ItemType {
	return TypeArtist
}

func ArtistsToItems(artists []*Artist) []Item {
	if artists == nil {
		return []Item{}
	}
	items := make([]Item, len(artists))

	for i, v := range artists {
		items[i] = v
	}
	return items
}
