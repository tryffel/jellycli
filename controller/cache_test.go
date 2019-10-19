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

package controller

import (
	"testing"
	"tryffel.net/pkg/jellycli/models"
)

func TestCache_Put(t *testing.T) {
	c, _ := NewCache()
	tests := []struct {
		name   string
		id     models.Id
		value  models.Item
		expire bool
	}{
		{
			id: "artist",
			value: &models.Artist{
				Id:            "a",
				Name:          "b",
				Albums:        nil,
				TotalDuration: 0,
			},
			expire: true,
		},
		{
			id: "album",
			value: &models.Album{
				Id:            "b",
				Name:          "c",
				Year:          2019,
				TotalDuration: 0,
				Artist:        "",
				Songs:         nil,
			},
			expire: true,
		},
		{
			id: "song",
			value: &models.Song{
				Id:     "song",
				Name:   "",
				Length: 180,
				Album:  "",
			},
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Put(tt.id, tt.value, tt.expire)
			val, found := c.Get(tt.id)
			if !found {
				t.Errorf("didn't found item from cache")
			}
			if val != tt.value {
				t.Errorf("cache value doesn't match original: want: %v, got: %v", tt.value, val)
			}
			if c.Count() != (i + 1) {
				t.Errorf("Cache count doesn't match, want %d, got %d", i+1, c.Count())
			}

		})
	}
}

func TestCache_Delete(t *testing.T) {
	c, _ := NewCache()
	tests := []struct {
		name   string
		id     models.Id
		value  models.Item
		expire bool
	}{
		{
			id: "id1",
			value: &models.Album{
				Id: "item1",
			},
			expire: true,
		},
		{
			id: "id2",
			value: &models.Artist{
				Id:   "item2",
				Name: "artist",
			},
			expire: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Put(tt.id, tt.value, tt.expire)
			c.Delete(tt.id)
			item, found := c.Get(tt.id)
			if item == tt.value || found {
				t.Errorf("deleted value was still found")
			}
		})
	}
}
