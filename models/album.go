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

type Album struct {
	Id            string
	Name          string
	Year          string
	TotalDuration int
	Songs         []*Song
}

func (a *Album) GetId() string {
	return a.Id
}

func (a *Album) HasChildren() bool {
	return len(a.Songs) > 0
}

func (a *Album) GetChildren() []Item {
	m := make([]Item, len(a.Songs))
	for i, v := range a.Songs {
		m[i] = v
	}
	return m
}

func (a *Album) GetParent() string {
	return ""
}

func (a *Album) GetName() string {
	return a.Name
}

func (a *Album) GetType() ListElement {
	return AlbumList
}
