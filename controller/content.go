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
	"github.com/sirupsen/logrus"
	"sync"
	"tryffel.net/pkg/jellycli/api"
	"tryffel.net/pkg/jellycli/models"
	"tryffel.net/pkg/jellycli/player"
	"tryffel.net/pkg/jellycli/task"
)

type Action int

const (
	ActionSearch Action = 0
)

type Content struct {
	task.Task
	api    *api.Api
	cache  *Cache
	player *player.Player
	lock   sync.RWMutex

	searchResults *api.SearchResult
	chanComplete  chan Action

	statusChangedCb func(state player.PlayingState)
	itemsCb         func([]models.Item)

	queue *queue
}

func (c *Content) GetChildren(id models.Id) {
	if c.itemsCb == nil {
		return
	}
	parent, found := c.cache.Get(id)
	if !found {
		return
	}
	children := c.getItems(parent.GetChildren())
	c.itemsCb(children)
	c.flushItems(children...)
}

func (c *Content) GetParent(id models.Id) {
	child, found := c.cache.Get(id)
	if !found {
		return
	}
	parentId := child.GetParent()
	parent := c.getItem(parentId)
	c.flushItems(parent)
}

func (c *Content) GetItem(id models.Id) {
	item := c.getItem(id)
	c.flushItems(item)
}

func (c *Content) GetItems(ids []models.Id) {
	if c.itemsCb == nil {
		return
	}
	items := c.getItems(ids)
	c.flushItems(items...)
}

func (c *Content) getItem(id models.Id) models.Item {
	if c.itemsCb == nil {
		return nil
	}
	item, found := c.cache.Get(id)
	if found {
		return item
	}
	return nil
}

func (c *Content) getItems(ids []models.Id) []models.Item {
	items := make([]models.Item, 0)
	for _, v := range ids {
		item, found := c.cache.Get(v)
		if found && item != nil {
			items = append(items, item)
		}
	}
	return items
}

func (c *Content) SetItemsCallback(cb func([]models.Item)) {
	c.itemsCb = cb
}

func (c *Content) RemoveItemsCallback() {
	c.itemsCb = nil
}

func (c *Content) GetQueue() []*models.Song {
	return c.queue.GetQueue()
}

func (c *Content) ClearQueue() {
	c.queue.ClearQueue()
}

func (c *Content) QueueDuration() int {
	return c.queue.QueueDuration()
}

func (c *Content) AddSongs(songs []*models.Song) {
	c.queue.AddSongs(songs)
}

func (c *Content) Reorder(currentIndex, newIndex int) {
	c.queue.Reorder(currentIndex, newIndex)
}

func (c *Content) GetHistory(n int) []*models.Song {
	return c.queue.GetHistory(n)
}

func (c *Content) SetQueueChangedCallback(cb func(content []*models.Song)) {
	c.queue.SetQueueChangedCallback(cb)
}

func (c *Content) RemoveQueueChangedCallback() {
	c.queue.RemoveQueueChangedCallback()
}

func (c *Content) Pause() {
	a := player.Action{
		State: player.Pause,
	}
	c.flushStatus(a)
}

func (c *Content) SetVolume(level int) {
	a := player.Action{
		State:  player.SetVolume,
		Volume: level,
	}
	c.flushStatus(a)

}

func (c *Content) Continue() {
	a := player.Action{
		State: player.Continue,
	}
	c.flushStatus(a)
}

func (c *Content) Stop() {
	a := player.Action{
		State: player.Stop,
	}
	c.flushStatus(a)
}

func (c *Content) Next() {
}

func (c *Content) Previous() {
}

func (c *Content) Seek(seconds int) {
}

func (c *Content) SeekBackwards(seconds int) {
}

func (c *Content) SetStatusCallback(cb func(state player.PlayingState)) {
	c.statusChangedCb = cb
}

func (c *Content) flushStatus(status player.Action) {
	if c.player != nil {
		c.player.ActionChannel() <- status
	}
}

//flushItems flushes either single item or multiple items
func (c *Content) flushItems(items ...models.Item) {
	if items != nil && c.itemsCb != nil {
		c.itemsCb(items)
	}
}

func NewContent(a *api.Api, cache *Cache) *Content {
	c := &Content{
		api:   a,
		cache: cache,
	}
	c.SetLoop(c.loop)
	c.chanComplete = make(chan Action)

	data := testData()
	for _, v := range data {
		c.cache.Put(v.GetId(), v, true)
	}

	return c
}

// Search performs search query
func (c *Content) Search(q string) {
	results, err := c.api.Search(q, 20)
	if err != nil {
		logrus.Error("Search failed: ", err.Error())
	} else {
		if results != nil {
			c.searchResults = results
		}
	}
	c.chanComplete <- ActionSearch
	logrus.Debug("Content search copmlete")

}

//SearchResults returns latest search results from index to index.
func (c *Content) SearchResults() *api.SearchResult {
	return c.searchResults
}

func (c *Content) SearchCompleteChan() chan Action {
	return c.chanComplete
}

func (c *Content) loop() {
	for true {
		select {
		case <-c.StopChan():
			break
		case state := <-c.player.StateChannel():
			if c.statusChangedCb != nil {
				c.statusChangedCb(state)
			}
		}
	}

}
