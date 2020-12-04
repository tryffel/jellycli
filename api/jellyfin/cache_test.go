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

package jellyfin

import (
	"reflect"
	"testing"
	"tryffel.net/go/jellycli/models"
)

func cacheTestData() []models.Item {
	return []models.Item{
		&models.Song{
			Id:       "s1",
			Name:     "song1",
			Duration: 0,
			Index:    0,
			Album:    "",
		},
		&models.Album{
			Id:       "a1",
			Name:     "album1",
			Year:     0,
			Duration: 0,
			Artist:   "",
			Songs:    nil,
		},
		&models.Artist{
			Id:            "ar1",
			Name:          "artist1",
			Albums:        nil,
			TotalDuration: 0,
		},
		&models.Artist{
			Id:            "ar2",
			Name:          "artist2",
			Albums:        nil,
			TotalDuration: 0,
		},
	}
}

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
				Id:       "b",
				Name:     "c",
				Year:     2019,
				Duration: 0,
				Artist:   "",
				Songs:    nil,
			},
			expire: true,
		},
		{
			id: "song",
			value: &models.Song{
				Id:       "song",
				Name:     "",
				Duration: 180,
				Album:    "",
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

func TestCache_PutBatch(t *testing.T) {
	items := cacheTestData()

	tests := []struct {
		name    string
		items   []models.Item
		wantErr bool
	}{
		{
			items:   items,
			wantErr: false,
		},
		{
			name: "missing id",
			items: append(items, &models.Album{
				Id:       "",
				Name:     "",
				Year:     0,
				Duration: 0,
				Artist:   "",
				Songs:    nil,
			}),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		cache, err := NewCache()
		if err != nil {
			t.Error("Failed to initialize cache: ", cache)
			return
		}
		cache.cache.Flush()
		t.Run(tt.name, func(t *testing.T) {
			if err = cache.PutBatch(tt.items, true); (err != nil) != tt.wantErr {
				t.Errorf("PutBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCache_GetBatch(t *testing.T) {
	items := cacheTestData()
	itemIds := make([]models.Id, len(items))
	for i, v := range items {
		itemIds[i] = v.GetId()
	}

	testItems := []models.Item{
		&models.Song{
			Id:       "song2",
			Name:     "",
			Duration: 0,
			Index:    0,
			Album:    "",
		},
		&models.Song{
			Id:       "song5",
			Name:     "",
			Duration: 0,
			Index:    0,
			Album:    "",
		},
	}

	tests := []struct {
		name         string
		items        []models.Item
		wantItems    []models.Item
		wantIds      []models.Id
		wantAllFound bool
	}{
		{
			name:         "items_found",
			items:        items,
			wantItems:    items,
			wantIds:      itemIds,
			wantAllFound: true,
		},
		{
			name:         "missing_item",
			items:        items,
			wantItems:    items,
			wantIds:      append(itemIds, "missing1"),
			wantAllFound: false,
		},
		{
			name:         "missing_items_unordered",
			items:        append(items, testItems...),
			wantIds:      append(itemIds, "missing2", "another-missing", "song2", "song5", "missing1"),
			wantItems:    append(items, testItems...),
			wantAllFound: false,
		},
	}
	for _, tt := range tests {
		cache, err := NewCache()
		if err != nil {
			t.Error("Failed to initialize cache: ", cache)
			return
		}
		t.Run(tt.name, func(t *testing.T) {
			cache.cache.Flush()

			err = cache.PutBatch(tt.items, true)
			if err != nil {
				t.Errorf("PutBatch() failed: %v", err)
			}
			foundItems, foundAll := cache.GetBatch(tt.wantIds)

			if foundAll != tt.wantAllFound {
				t.Errorf("GetBatch() foundAll = %v, wantFoundAll %v", foundAll, tt.wantAllFound)
			}

			if !reflect.DeepEqual(foundItems, tt.wantItems) {
				t.Errorf("GetBatch() gotItems = %v, wantItems %v", items, tt.wantItems)
			}
		})
	}
}
