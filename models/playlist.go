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
