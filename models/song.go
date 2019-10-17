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
	Id     string
	Name   string
	Length int
}

func (s *Song) GetId() string {
	return s.Id
}

func (s *Song) HasChildren() bool {
	return false
}

func (s *Song) GetChildren() []Item {
	return []Item{}
}

func (s *Song) GetParent() string {
	return ""
}

func (s *Song) GetName() string {
	return s.Name
}

func (s *Song) GetType() ListElement {
	return SongList
}
