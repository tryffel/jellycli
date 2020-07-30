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

// Artist has multiple albums.
type Artist struct {
	Id            Id
	Name          string
	Albums        []Id
	TotalDuration int
	AlbumCount    int

	Favorite bool
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
