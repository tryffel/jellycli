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

package player

import (
	"github.com/sirupsen/logrus"
	"runtime"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

// Items implements interfaces.ItemController
type Items struct {
	browser api.MediaServer
}

func newItems(api api.MediaServer) *Items {
	return &Items{
		browser: api,
	}
}

func (i *Items) Search(itemType models.ItemType, query string) ([]models.Item, error) {
	return i.browser.Search(query, itemType, config.AppConfig.Player.SearchResultsLimit)
}

func (i *Items) GetArtists(paging interfaces.Paging) ([]*models.Artist, int, error) {
	return i.browser.GetArtists(paging)
}

func (i *Items) GetAlbumArtists(paging interfaces.Paging) ([]*models.Artist, int, error) {
	return i.browser.GetAlbumArtists(paging)
}

func (i *Items) GetAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	return i.browser.GetAlbums(paging)
}

func (i *Items) GetArtistAlbums(artist models.Id) ([]*models.Album, error) {
	return i.browser.GetArtistAlbums(artist)
}

func (i *Items) GetAlbumSongs(album models.Id) ([]*models.Song, error) {
	return i.browser.GetAlbumSongs(album)
}

func (i *Items) GetPlaylists() ([]*models.Playlist, error) {
	return i.browser.GetPlaylists()
}

func (i *Items) GetPlaylistSongs(playlist *models.Playlist) error {
	songs, err := i.browser.GetPlaylistSongs(playlist.Id)
	if err != nil {
		return err
	}
	playlist.Songs = songs

	return nil
}

func (i *Items) GetFavoriteArtists() ([]*models.Artist, error) {
	return i.browser.GetFavoriteArtists()
}

func (i *Items) GetFavoriteAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	return i.browser.GetFavoriteAlbums(paging)
}

func (i *Items) GetLatestAlbums() ([]*models.Album, error) {
	return i.browser.GetLatestAlbums()
}

func (i *Items) GetRecentlyPlayed(paging interfaces.Paging) ([]*models.Song, int, error) {
	return i.browser.GetRecentlyPlayed(paging)
}

func (i *Items) GetSimilarArtists(artist models.Id) ([]*models.Artist, error) {
	return i.browser.GetSimilarArtists(artist)
}

func (i *Items) GetSimilarAlbums(album models.Id) ([]*models.Album, error) {
	return i.browser.GetSimilarAlbums(album)
}

func (i *Items) GetGenres(paging interfaces.Paging) ([]*models.IdName, int, error) {
	return i.browser.GetGenres(paging)
}

func (i *Items) GetGenreAlbums(genre models.IdName) ([]*models.Album, error) {
	return i.browser.GetGenreAlbums(genre)
}

func (i *Items) GetStatistics() models.Stats {
	runStats := runtime.MemStats{}
	runtime.ReadMemStats(&runStats)

	stats := models.Stats{
		Heap:       int(runStats.Alloc),
		LogFile:    config.LogFile,
		ConfigFile: config.ConfigFile,
	}

	var err error
	stats.ServerInfo, err = i.browser.GetInfo()
	if err != nil {
		logrus.Errorf("get server info: %v", err)
	}
	return stats
}

func (i *Items) GetSongs(page, pageSize int) ([]*models.Song, int, error) {
	return i.browser.GetSongs(page, pageSize)
}

func (i *Items) GetAlbumArtist(album *models.Album) (*models.Artist, error) {
	return i.browser.GetAlbumArtist(album)
}

func (i *Items) GetSongArtistAlbum(song *models.Song) (*models.Album, *models.Artist, error) {
	return i.browser.GetSongArtistAlbum(song)
}

func (i *Items) GetInstantMix(item models.Item) ([]*models.Song, error) {
	return i.browser.GetInstantMix(item)
}

func (i *Items) GetLink(item models.Item) string {
	return i.browser.GetLink(item)

}
