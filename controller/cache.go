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
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/models"
)

type Cache struct {
	cache *cache.Cache
}

func NewCache() (*Cache, error) {
	c := &Cache{}
	c.cache = cache.New(config.CacheTimeout, config.CacheTimeout*2)
	return c, nil
}

func (c *Cache) Count() int {
	return c.cache.ItemCount()
}

func (c *Cache) Put(id models.Id, item models.Item, expire bool) {
	var timeout time.Duration
	if expire {
		timeout = config.CacheTimeout
	} else {
		timeout = cache.NoExpiration
	}
	c.cache.Set(string(id), item, timeout)
}

func (c *Cache) Get(id models.Id) (models.Item, bool) {
	count := c.Count()
	logrus.Debugf("Cache has %d items", count)

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

func (c *Cache) Delete(id models.Id) {
	c.cache.Delete(string(id))
}
