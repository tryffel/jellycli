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
	"time"
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

	chanItemsAdded chan []*models.Song

	statusChangedCb func(state player.PlayingState)
	itemsCb         func([]models.Item)

	playerState player.PlayingState

	ticker *time.Ticker
	queue  *queue
}

func (c *Content) GetChildren(parent models.Id, parentType models.ItemType) {
	if c.itemsCb == nil {
		return
	}
	var items []models.Item

	switch parentType {
	case models.TypeArtist:
		albums, err := c.api.GetArtistAlbums(parent)
		if err != nil {
			logrus.Errorf("failed to get artist albums: %v", err)
			return
		}
		items = make([]models.Item, len(albums))
		for i, v := range albums {
			items[i] = v
		}
	case models.TypeAlbum:
		songs, err := c.api.GetAlbumSongs(parent)
		if err != nil {
			logrus.Errorf("failed to get album songs: %v", err)
		}
		items = make([]models.Item, len(songs))
		for i, v := range songs {
			items[i] = v
		}
	case models.TypeSong:
		logrus.Debug("Song has no child")
		return
	default:
		logrus.Warningf("Invalid request for getChildren() for type %s", parentType)
		return
	}

	c.flushItems(items...)
}

func (c *Content) GetParent(child models.Id, childType models.ItemType) {
	/*
		if c.itemsCb == nil {
			return
		}
		var item models.Item
		var err error

		switch childType {
		case models.TypeArtist:
			logrus.Debug("Artist has no parent")
			return
		case models.TypeAlbum:
			artist, err := c.api.GetArtist(child)
			if err != nil {
				logrus.Errorf("failed to get album songs: %v", err)
				return
			}

			item = &artist
		case models.TypeSong:
			album, err := c.api.getalbu
		default:
			logrus.Warningf("Invalid request for getChildren() for type %s", parentType)
			return
		}

		c.flushItems(items...)
	*/
}

func (c *Content) GetItem(id models.Id, itemType models.ItemType) {
	if c.itemsCb == nil {
		logrus.Debug("No itemsCallback set, not getting item")
		return
	}
	var item models.Item

	switch itemType {
	case models.TypeArtist:
		artist, err := c.api.GetArtist(id)
		if err != nil {
			logrus.Errorf("failed to get artist: %v", err)
			return
		}
		item = &artist
	case models.TypeAlbum:
		songs, err := c.api.GetAlbum(id)
		if err != nil {
			logrus.Errorf("failed to get album: %v", err)
			return
		}
		item = &songs
	case models.TypeSong:
		logrus.Errorf("Cannot get item(song, %s) from api", id)
		return
	default:
		logrus.Warningf("Invalid type to get item for: %s", itemType)
		return
	}

	c.flushItems(item)
}

func (c *Content) GetItems(ids []models.Id, itemType models.ItemType) {
	if c.itemsCb == nil {
		logrus.Debug("No itemsCallback set, not getting items")
		return
	}
	logrus.Errorf("GetItems not implemented")
	return
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
	c.chanItemsAdded <- songs
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
		queue:  newQueue(),
	}

	c.SetLoop(c.loop)
	c.chanComplete = make(chan Action)
	c.chanItemsAdded = make(chan []*models.Song)
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
	c.ticker = time.NewTicker(time.Second * 3)
	for true {
		select {
		case <-c.StopChan():
			break
		case state := <-c.player.StateChannel():
			c.playerState = state

			if state.State == player.SongComplete {
				c.queue.songComplete()
			}

			if c.statusChangedCb != nil {
				c.statusChangedCb(state)
			}
		case songs := <-c.chanItemsAdded:
			c.queue.AddSongs(songs)
			c.ensurePlayerHasStream()
		case <-c.ticker.C:
			c.ticker.Stop()
			c.ensurePlayerHasStream()
			c.ticker = time.NewTicker(time.Second * 3)
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

func (c *Content) ensurePlayerHasStream() {
	// Ensure player has something to play.
	if c.queue.empty() {
		return
	}

	// Download new song if current song is almost finished
	//left := c.playerState.CurrentSongDuration - c.playerState.CurrentSongPast
	//if left < 10 {

	//}

	if c.playerState.State == player.Stop || c.playerState.State == player.SongComplete {
		song := c.queue.currentSong()
		logrus.Debugf("Giving player a new song to play: %s", song.Name)

		albumId := song.GetParent()
		album, err := c.api.GetAlbum(albumId)
		artist := models.Artist{Name: "unknown artist"}
		if err != nil {
			logrus.Error("Failed to get album by id: %v", err)
			album = models.Album{Name: "unknown album"}
		} else {
			a, err := c.api.GetArtist(album.GetParent())
			if err != nil {
				logrus.Errorf("Failed to get artist by id: %v", err)
			} else {
				artist = a
			}
		}

		action := player.Action{
			State:    player.Play,
			Type:     player.Song,
			Volume:   0,
			Artist:   artist.Name,
			Album:    album.Name,
			Song:     song.Name,
			Year:     album.Year,
			AudioId:  song.Id.String(),
			Duration: song.Duration,
		}
		c.player.PlaySong(action)
	}
}
