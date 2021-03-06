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
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/storage"
)

// Items implements interfaces.ItemController
type Items struct {
	browser api.MediaServer

	db *storage.Db
}

func newItems(api api.MediaServer) (*Items, error) {
	items := &Items{
		browser: api,
	}
	var err error

	serverId := api.GetId()
	if config.AppConfig.Player.EnableLocalCache {
		items.db, err = storage.NewDb(serverId)
		if err != nil {
			return items, fmt.Errorf("init local database: %v", err)
		}
	}
	return items, err
}

func (i *Items) closeDb() {
	if i.db != nil {
		err := i.db.Close()
		if err != nil {
			logrus.Errorf("close db: %s", err)
		}
	}
}

func (i *Items) Search(itemType models.ItemType, query string) ([]models.Item, error) {
	return i.browser.Search(query, itemType, config.AppConfig.Gui.SearchResultsLimit)
}

func (i *Items) GetArtists(opts *interfaces.QueryOpts) ([]*models.Artist, int, error) {
	if config.AppConfig.Player.EnableLocalCache {
		return i.db.GetArtists(opts)
	} else {
		return i.browser.GetArtists(opts)
	}
}

func (i *Items) GetAlbumArtists(paging interfaces.Paging) ([]*models.Artist, int, error) {
	return i.browser.GetAlbumArtists(interfaces.DefaultQueryOpts())
}

func (i *Items) GetAlbums(opts *interfaces.QueryOpts) ([]*models.Album, int, error) {
	if config.AppConfig.Player.EnableLocalCache {
		return i.db.GetAlbums(opts)
	} else {
		return i.browser.GetAlbums(opts)
	}
}

func (i *Items) GetArtistAlbums(artist models.Id) ([]*models.Album, error) {
	return i.browser.GetArtistAlbums(artist)
}

func (i *Items) GetAlbumSongs(album models.Id) ([]*models.Song, error) {
	return i.browser.GetAlbumSongs(album)
}

func (i *Items) GetPlaylists() ([]*models.Playlist, error) {
	if config.AppConfig.Player.EnableLocalCache {
		return i.db.GetPlaylists()
	} else {
		return i.browser.GetPlaylists()
	}

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
	query := interfaces.DefaultQueryOpts()
	query.Filter.Favorite = true
	artists, _, err := i.browser.GetArtists(query)
	return artists, err
}

func (i *Items) GetFavoriteAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	query := interfaces.DefaultQueryOpts()
	query.Filter.Favorite = true
	query.Paging = paging
	return i.browser.GetAlbums(query)
}

func (i *Items) GetLatestAlbums() ([]*models.Album, error) {
	query := interfaces.DefaultQueryOpts()
	if config.AppConfig.Gui.LimitRecentlyPlayed {
		query.Paging.PageSize = 100
	}
	query.Sort.Field = interfaces.SortByLatest
	query.Sort.Mode = interfaces.SortDesc
	albums, _, err := i.browser.GetAlbums(query)
	return albums, err
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
	query := interfaces.DefaultQueryOpts()
	query.Filter.Genres = []models.IdName{genre}

	albums, _, err := i.browser.GetAlbums(query)
	return albums, err
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

	if i.db != nil {
		stats.StorageInfo, err = i.db.GetStats()
		if err != nil {
			logrus.Errorf("get local storage info: %v", err)
		}
	}
	return stats
}

func (i *Items) GetSongs(page, pageSize int) ([]*models.Song, int, error) {
	if config.AppConfig.Player.EnableLocalCache {
		return i.db.GetSongs(page, pageSize)
	} else {
		return i.browser.GetSongs(interfaces.DefaultQueryOpts())
	}
}

func (i *Items) GetAlbumArtist(album *models.Album) (*models.Artist, error) {
	return i.browser.GetAlbumArtist(album)
}

func (i *Items) GetSongArtistAlbum(song *models.Song) (*models.Album, *models.Artist, error) {
	id := song.AlbumArtist
	if id == "" && len(song.Artists) > 0 {
		id = song.Artists[0].Id
	}

	artist, err := i.browser.GetArtist(id)
	if err != nil {
		return nil, artist, err
	}
	album, err := i.browser.GetAlbum(song.Album)
	return album, artist, err
}

func (i *Items) GetInstantMix(item models.Item) ([]*models.Song, error) {
	return i.browser.GetInstantMix(item)
}

func (i *Items) GetLink(item models.Item) string {
	return i.browser.GetLink(item)

}
