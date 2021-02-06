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

import (
	"database/sql/driver"
	"fmt"
)

type Id string

func (i Id) Value() (driver.Value, error) {
	return string(i), nil
}

func (i *Id) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("'%v' not string", src)
	}
	*i = Id(str)
	return nil
}

func (i Id) String() string {
	return string(i)
}

// Item is any object that has unique id and falls to some category with ItemType.
type Item interface {
	GetId() Id
	GetName() string
	HasChildren() bool
	GetChildren() []Id
	GetParent() Id
	GetType() ItemType
}

type ItemType string

const (
	TypeArtist   ItemType = "Artist"
	TypeAlbum    ItemType = "Album"
	TypePlaylist ItemType = "Playlist"
	TypeQueue    ItemType = "Queue"
	TypeHistory  ItemType = "History"
	TypeSong     ItemType = "Song"
	TypeGenre    ItemType = "Genre"
)
