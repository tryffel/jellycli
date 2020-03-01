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
	"runtime"
	"sync"
	"time"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/player"
	"tryffel.net/go/jellycli/task"
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

	statusChangedCb []func(state interfaces.PlayingState)
	itemsCb         func([]models.Item)

	playerState interfaces.PlayingState

	ticker *time.Ticker
	queue  *queue
	// is new song being downloaded
	downloadingSong bool
	// is new song pending acknowledgment from player
	newSongPending bool
}

//is download pending / ongoing
func (c *Content) isDownloadingSong() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.downloadingSong
}

func (c *Content) isNewSongPending() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.newSongPending
}

func (c *Content) SearchArtists(search string) ([]*models.Artist, error) {
	panic("implement me")
}

func (c *Content) SearchAlbums(search string) ([]*models.Album, error) {
	panic("implement me")
}

func (c *Content) GetArtists() ([]*models.Artist, error) {
	panic("implement me")
}

func (c *Content) GetLatestAlbums() ([]*models.Album, error) {
	return c.api.GetLatestAlbums()
}

func (c *Content) GetAlbums() ([]*models.Album, error) {
	panic("implement me")
}

func (c *Content) GetAlbumSongs(album models.Id) ([]*models.Song, error) {
	return c.api.GetAlbumSongs(album)
}

func (c *Content) GetPlaylists() ([]*models.Album, error) {
	panic("implement me")
}

func (c *Content) GetFavoriteArtists() ([]*models.Artist, error) {
	return c.api.GetFavoriteArtists()
}

func (c *Content) GetFavoriteAlbums() ([]*models.Album, error) {
	panic("implement me")
}

func (c *Content) GetArtistAlbums(artist models.Id) ([]*models.Album, error) {
	return c.api.GetArtistAlbums(artist)
}

func (c *Content) GetStatistics() models.Stats {
	cache := c.api.GetCacheItems()
	name, version, _, _ := c.api.GetServerVersion()
	runStats := runtime.MemStats{}
	runtime.ReadMemStats(&runStats)

	stats := models.Stats{
		Heap:          int(runStats.Alloc),
		CacheObjects:  cache,
		ServerName:    name,
		ServerVersion: version,
		WebSocket:     c.api.WebSocketEnabled(),
		LogFile:       config.LogFile,
	}

	return stats
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

func (c *Content) GetView(view interfaces.View) {
	var err error
	var items []models.Item

	switch view {
	case interfaces.ViewAllArtists:
	case interfaces.ViewAllAlbums:
	case interfaces.ViewAllSongs:
	case interfaces.ViewFavoriteArtists:
	case interfaces.ViewFavoriteAlbums:
	case interfaces.ViewFavoriteSongs:
	case interfaces.ViewPlaylists:
	case interfaces.ViewLatestMusic:
		var albums []*models.Album
		albums, err = c.api.GetLatestAlbums()
		if err == nil {
			items = models.AlbumsToItems(albums)
		}
	}

	if err == nil && items != nil {
		if len(items) > 0 {
			c.flushItems(items...)
		}
	} else if err != nil {
		logrus.Errorf("failed to get view %d: %v", view, err)
	}

}

func (c *Content) RemoveItemsCallback() {
	c.itemsCb = nil
}

func (c *Content) GetQueue() []*models.SongInfo {

	return c.songsToInfos(c.queue.GetQueue())
}

func (c *Content) songsToInfos(songs []*models.Song) []*models.SongInfo {
	now := time.Now()
	requests := 0
	infos := make([]*models.SongInfo, len(songs))
	// This shouldn't take that long since user has added each item just now and thus each item should exist
	// in cache
	//TODO: use single or two requests for all items
	for i, v := range songs {
		infos[i] = songs[i].ToInfo()
		album, err := c.api.GetAlbum(v.Album)
		requests += 1
		if err != nil {
			logrus.Warning("Failed to get album for song: ", err)
		} else {
			infos[i].Album = album.Name
			infos[i].Year = album.Year
			requests += 1
			artist, err := c.api.GetArtist(album.Artist)
			if err != nil {
				logrus.Error("Failed to get artist for album: ", err)
			} else {
				infos[i].Artist = artist.Name
			}
		}
	}

	took := time.Since(now)
	logrus.Debugf("Gathering queue info took %d ms with %d requests", took.Milliseconds(), requests)
	return infos
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

func (c *Content) GetHistory(n int) []*models.SongInfo {
	return c.songsToInfos(c.queue.GetHistory(n))
}

func (c *Content) SetQueueChangedCallback(cb func(content []*models.Song)) {
	c.queue.SetQueueChangedCallback(cb)
}

func (c *Content) RemoveQueueChangedCallback() {
	c.queue.RemoveQueueChangedCallback()
}

func (c *Content) Pause() {
	a := player.Action{
		State: interfaces.Pause,
	}
	c.flushStatus(a)
}

func (c *Content) PlayPause() {
	a := player.Action{}

	if c.playerState.State == interfaces.Play {
		a.State = interfaces.Pause
	} else if c.playerState.State == interfaces.Pause {
		a.State = interfaces.Continue
	}
	c.flushStatus(a)
}

func (c *Content) SetVolume(level int) {
	a := player.Action{
		State:  interfaces.SetVolume,
		Volume: level,
	}
	c.flushStatus(a)

}

func (c *Content) Continue() {
	a := player.Action{
		State: interfaces.Continue,
	}
	c.flushStatus(a)
}

func (c *Content) StopMedia() {
	c.queue.ClearQueue()
	a := player.Action{
		State: interfaces.Stop,
	}
	c.flushStatus(a)
}

func (c *Content) Next() {
	if len(c.queue.GetQueue()) > 1 {
		// don't skip track if there's no more tracks available
		status := player.Action{
			State: interfaces.EndSong,
		}
		c.flushStatus(status)
	}
}

func (c *Content) Previous() {
}

func (c *Content) Seek(seconds int) {
}

func (c *Content) SeekBackwards(seconds int) {
}

func (c *Content) AddStatusCallback(cb func(status interfaces.PlayingState)) {
	c.statusChangedCb = append(c.statusChangedCb, cb)
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
		api:             a,
		player:          p,
		queue:           newQueue(),
		statusChangedCb: []func(tate interfaces.PlayingState){},
	}

	c.SetLoop(c.loop)
	c.chanComplete = make(chan Action, 3)
	c.chanItemsAdded = make(chan []*models.Song, 3)
	c.Name = "Content"
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
			if state.State == interfaces.SongStarted {
				if c.isNewSongPending() {
					c.lock.Lock()
					c.newSongPending = false
					c.lock.Unlock()
				}
			}
			c.playerState = state

			err := c.pushState(state)
			if err != nil {
				logrus.Errorf("push status: %v", err)
			}
			if state.State == interfaces.SongComplete {
				c.queue.songComplete()
				c.ensurePlayerHasStream()
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

func (c *Content) pushState(status interfaces.PlayingState) error {
	if status.State != interfaces.Stop {
		if status.Album != nil && status.AlbumImageUrl == "" {
			status.AlbumImageUrl = c.api.ImageUrl(string(status.Album.Id), status.Album.ImageId)
		}
	}

	for _, v := range c.statusChangedCb {
		v(status)
	}
	return nil
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

	if c.isDownloadingSong() || c.isNewSongPending() {
		return
	}

	// Download new song if current song is almost finished
	//left := c.playerState.CurrentSongDuration - c.playerState.CurrentSongPast
	//if left < 10 {

	//}

	if c.playerState.State == interfaces.Stop || c.playerState.State == interfaces.SongComplete {
		song := c.queue.currentSong()
		logrus.Debugf("Giving player a new song to play: %s", song.Name)

		albumId := song.GetParent()
		album, err := c.api.GetAlbum(albumId)
		artist := models.Artist{Name: "unknown artist"}
		if err != nil {
			logrus.Error("Failed to get album by id: ", err.Error())
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
			State:    interfaces.Play,
			Type:     interfaces.Song,
			Volume:   0,
			Artist:   &artist,
			Album:    &album,
			Song:     song,
			AudioId:  song.Id.String(),
			Duration: song.Duration,
		}
		go c.getNextSong(action)
	}
}

func (c *Content) getNextSong(action player.Action) {
	c.lock.Lock()
	c.downloadingSong = true
	c.lock.Unlock()

	reader, err := c.api.GetSongDirect(action.AudioId, "mp3")
	if err != nil {
		logrus.Error("failed to request file over http: ", err.Error())
	} else {
		song := player.PlaySong{
			Action: action,
			Song:   reader,
		}
		c.player.AddSong(song)
	}
	c.lock.Lock()
	c.downloadingSong = false
	c.newSongPending = true
	c.lock.Unlock()
}
