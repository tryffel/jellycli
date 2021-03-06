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
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
)

type Cache struct {
	cache *cache.Cache
}

//NewCache creates new cache that's ready to use.
func NewCache() (*Cache, error) {
	c := &Cache{}
	c.cache = cache.New(config.CacheTimeout, config.CacheTimeout*2)
	return c, nil
}

//Count returns total count of stored items
func (c *Cache) Count() int {
	return c.cache.ItemCount()
}

//Put puts single item. If expire is true, item expires after default expiration
func (c *Cache) Put(id models.Id, item models.Item, expire bool) {
	var timeout time.Duration
	if expire {
		timeout = config.CacheTimeout
	} else {
		timeout = cache.NoExpiration
	}
	c.cache.Set(string(id), item, timeout)
}

//Get gets single item fro cache. Returns item and flag whether item is found
func (c *Cache) Get(id models.Id) (models.Item, bool) {
	data, found := c.cache.Get(string(id))
	if !found {
		return nil, false
	}
	item, ok := data.(models.Item)
	if !ok {
		logrus.Errorf("Cached item not models.Item: %v", data)
		c.Delete(id)
		return nil, false
	}
	return item, true
}

//Delete deletes item with given id. If item is not found, do nothing.
func (c *Cache) Delete(id models.Id) {
	c.cache.Delete(string(id))
}

//PutBatch put's multiple items with expiration. Each item must have a valid id
//or operation fails returning error.
func (c *Cache) PutBatch(items []models.Item, expire bool) error {
	for i, v := range items {
		id := v.GetId()
		if id == "" {
			return fmt.Errorf("%d item has no id", i)
		}
		c.Put(id, v, expire)
	}
	return nil
}

//GetBatch returns batch of items with given ids.
//Return array is always same length of ids. However, if not all items are found,
//return flag is set to false.
func (c *Cache) GetBatch(ids []models.Id) ([]models.Item, bool) {
	count := len(ids)
	foundTotal := 0
	items := make([]models.Item, count)

	for _, v := range ids {
		item, found := c.Get(v)
		if found {
			items[foundTotal] = item
			foundTotal += 1
		}
	}

	if count == foundTotal {
		return items, true
	}

	items = items[:foundTotal]
	return items, false
}

//PutList puts a list of ids under key
func (c *Cache) PutList(id string, data []models.Id) {
	c.cache.Set(id, data, config.CacheTimeout)
}

//GetList gets list of Ids with given id
func (c *Cache) GetList(id string) ([]models.Id, bool) {
	items, found := c.cache.Get(id)
	if !found {
		return nil, false
	}
	ids, ok := items.([]models.Id)
	if !ok {
		c.Delete(models.Id(id))
		return nil, false
	}
	return ids, true
}

func (c *Cache) GetArtist(id models.Id) *models.Artist {
	item, found := c.Get(id)
	if found {
		artist, ok := item.(*models.Artist)
		if ok {
			return artist
		}
	}
	return nil
}

func (c *Cache) GetAlbum(id models.Id) *models.Album {
	item, found := c.Get(id)
	if found {
		album, ok := item.(*models.Album)
		if ok {
			return album
		}
	}
	return nil
}

func (c *Cache) GetSong(id models.Id) *models.Song {
	item, found := c.Get(id)
	if found {
		song, ok := item.(*models.Song)
		if ok {
			return song
		}
	}
	return nil
}

func (c *Cache) GetPlaylist(id models.Id) *models.Playlist {
	item, found := c.Get(id)
	if found {
		playlist, ok := item.(*models.Playlist)
		if ok {
			return playlist
		}
	}
	return nil
}
