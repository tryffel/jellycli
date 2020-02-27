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

type Song struct {
	Id         Id
	Name       string
	Duration   int
	Index      int
	Album      Id
	DiscNumber int
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
