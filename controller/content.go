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
	"fmt"
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
	items, err := c.api.GetChildItems(id)
	if err != nil {
		logrus.Errorf("Failed to get children for %s: %v", id, err)
	} else if items != nil {
		c.flushItems(items...)
	}
}

func (c *Content) GetParent(id models.Id) {
	parent, err := c.api.GetParentItem(id)
	if err != nil {
		logrus.Errorf("Failed to get parent for %s: %v", id, err)
	} else if parent != nil {
		c.flushItems(parent)
	}
}

func (c *Content) GetItem(id models.Id) {
	if c.itemsCb == nil {
		return
	}
	item := c.getItem(id)
	c.flushItems(item)
}

func (c *Content) GetItems(ids []models.Id) {
	if c.itemsCb == nil || ids == nil {
		return
	}
	items := c.getItems(ids)
	c.flushItems(items...)
}

func (c *Content) getItem(id models.Id) models.Item {
	item, err := c.api.GetItem(id)
	if err != nil {
		logrus.Errorf("Failed to get item %s: %v", id, err)
		return nil
	}
	return item
}

func (c *Content) getItems(ids []models.Id) []models.Item {
	items, err := c.api.GetItems(ids)
	if err != nil {
		logrus.Errorf("Failed to multiple items (%d), first item: %s: %s", len(ids), ids[0], err)
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

func (c *Content) StopMedia() {
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

func NewContent(a *api.Api, p *player.Player) (*Content, error) {
	var err error
	c := &Content{
		api:    a,
		player: p,
	}

	c.SetLoop(c.loop)
	c.chanComplete = make(chan Action)
	if err != nil {
		return c, fmt.Errorf("init media player: %v", err)
	}
	return c, nil
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

func (c *Content) GetDefault() []models.Item {
	artists, err := c.api.GetFavoriteArtists()
	if err != nil {
		logrus.Error("Failed to retrieve favorite artists:", err)
		return nil
	}

	items := make([]models.Item, len(artists))
	for i, v := range artists {
		items[i] = v
	}

	return items
}
