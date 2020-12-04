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

// Playlist is a list of songs. It has no artists itself, but songs do have albums and artists.
type Playlist struct {
	Id       Id
	Name     string
	Duration int

	Songs     []*Song
	SongCount int
}

func (p Playlist) GetId() Id {
	return p.Id
}

func (p Playlist) GetName() string {
	return p.Name
}

func (p Playlist) HasChildren() bool {
	return p.SongCount > 0
}

func (p Playlist) GetChildren() []Id {
	ids := make([]Id, len(p.Songs))
	for i, v := range p.Songs {
		ids[i] = v.Id
	}
	return ids
}

func (p Playlist) GetParent() Id {
	return ""
}

func (p Playlist) GetType() ItemType {
	return TypePlaylist
}
