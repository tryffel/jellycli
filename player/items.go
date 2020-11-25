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
	"tryffel.net/go/jellycli/api/jellyfin"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

// Items implements interfaces.ItemController
type Items struct {
	api *jellyfin.Api
}

func newItems(api *jellyfin.Api) *Items {
	return &Items{
		api: api,
	}
}

func (i *Items) Search(itemType models.ItemType, query string) ([]models.Item, error) {
	return i.api.Search(query, itemType, config.AppConfig.Player.SearchResultsLimit)
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

func (i *Items) GetFavoriteAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	return i.api.GetFavoriteAlbums(paging)
}

func (i *Items) GetLatestAlbums() ([]*models.Album, error) {
	return i.api.GetLatestAlbums()
}

func (i *Items) GetRecentlyPlayed(paging interfaces.Paging) ([]*models.Song, int, error) {
	return i.api.GetRecentlyPlayed(paging)
}

func (i *Items) GetSimilarArtists(artist models.Id) ([]*models.Artist, error) {
	return i.api.GetSimilarArtists(artist)
}

func (i *Items) GetSimilarAlbums(album models.Id) ([]*models.Album, error) {
	return i.api.GetSimilarAlbums(album)
}

func (i *Items) GetGenres(paging interfaces.Paging) ([]*models.IdName, int, error) {
	return i.api.GetGenres(paging)
}

func (i *Items) GetGenreAlbums(genre models.IdName) ([]*models.Album, error) {
	return i.api.GetGenreAlbums(genre)
}

func (i *Items) GetStatistics() models.Stats {
	cache := i.api.GetCacheItems()
	name, version, id, restart, shutdown, _ := i.api.GetServerVersion()
	runStats := runtime.MemStats{}
	runtime.ReadMemStats(&runStats)

	stats := models.Stats{
		Heap:                  int(runStats.Alloc),
		CacheObjects:          cache,
		ServerName:            name,
		ServerVersion:         version,
		ServerId:              id,
		ServerRestartPending:  restart,
		ServerShutdownPending: shutdown,
		WebSocket:             i.api.WebsocketOk(),
		RemoteControl:         config.AppConfig.Player.EnableRemoteControl,
		LogFile:               config.LogFile,
		ConfigFile:            config.ConfigFile,
	}

	return stats
}

func (i *Items) GetSongs(page, pageSize int) ([]*models.Song, int, error) {
	return i.api.GetSongs(page, pageSize)
}

func (i *Items) GetAlbumArtist(album *models.Album) (*models.Artist, error) {
	return i.api.GetAlbumArtist(album)
}

func (i *Items) GetSongArtistAlbum(song *models.Song) (*models.Album, *models.Artist, error) {
	return i.api.GetSongArtistAlbum(song)
}

func (i *Items) GetInstantMix(item models.Item) ([]*models.Song, error) {
	return i.api.GetInstantMix(item)
}

func (i *Items) GetLink(item models.Item) string {
	return i.api.GetLink(item)

}
