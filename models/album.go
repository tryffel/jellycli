/*
 * Copyright 2019 Tero Vierimaa
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

package models

// Album has multiple songs. It has one primary artist and multiple additional artists.
type Album struct {
	Id       Id
	Name     string
	Year     int
	Duration int
	// Artist is the primary artist
	Artist Id
	// Additional artists. If length is 1, first item is same as primary artist. Else it contains additional artists
	AdditionalArtists []IdName
	Songs             []Id
	//SongCount, how many songs are there in album.
	// 0 means album is empty, where -1 means songs need to be gathered separately.
	SongCount int
	// ImageId is optional id for image album
	ImageId   string
	DiscCount int

	Favorite bool
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
