/*
 * Copyright 2020 Tero Vierimaa
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

package player

import (
	"runtime"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

// Items implements interfaces.ItemController
type Items struct {
	api *api.Api
}

func newItems(api *api.Api) *Items {
	return &Items{
		api: api,
	}
}

func (i *Items) SearchArtists(search string) ([]*models.Artist, error) {
	panic("implement me")
}

func (i *Items) SearchAlbums(search string) ([]*models.Album, error) {
	panic("implement me")
}

func (i *Items) GetArtists(paging interfaces.Paging) ([]*models.Artist, int, error) {
	return i.api.GetArtists(paging)
}

func (i *Items) GetAlbumArtists(paging interfaces.Paging) ([]*models.Artist, int, error) {
	return i.api.GetAlbumArtists(paging)
}

func (i *Items) GetAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	return i.api.GetAlbums(paging)
}

func (i *Items) GetArtistAlbums(artist models.Id) ([]*models.Album, error) {
	return i.api.GetArtistAlbums(artist)
}

func (i *Items) GetAlbumSongs(album models.Id) ([]*models.Song, error) {
	return i.api.GetAlbumSongs(album)
}

func (i *Items) GetPlaylists() ([]*models.Playlist, error) {
	return i.api.GetPlaylists()
}

func (i *Items) GetPlaylistSongs(playlist *models.Playlist) error {
	songs, err := i.api.GetPlaylistSongs(playlist.Id)
	if err != nil {
		return err
	}
	playlist.Songs = songs
	return nil
}

func (i *Items) GetFavoriteArtists() ([]*models.Artist, error) {
	return i.api.GetFavoriteArtists()
}

func (i *Items) GetFavoriteAlbums() ([]*models.Album, error) {
	panic("implement me")
}

func (i *Items) GetLatestAlbums() ([]*models.Album, error) {
	return i.api.GetLatestAlbums()
}

func (i *Items) GetStatistics() models.Stats {
	cache := i.api.GetCacheItems()
	name, version, _, _ := i.api.GetServerVersion()
	runStats := runtime.MemStats{}
	runtime.ReadMemStats(&runStats)

	stats := models.Stats{
		Heap:          int(runStats.Alloc),
		CacheObjects:  cache,
		ServerName:    name,
		ServerVersion: version,
		WebSocket:     i.api.WebsocketOk(),
		LogFile:       config.LogFile,
	}

	return stats
}

func (i *Items) GetSongs(page, pageSize int) ([]*models.Song, int, error) {
	return i.api.GetSongs(page, pageSize)
}
