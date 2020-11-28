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

// Package api contains interface for connecting to remote server. Subpackages contain implementations.
package api

import (
	"io"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

// MediaServer combines minimal interfaces for browsing and playing songs from remote server.
// Mediaserver can additionally implement RemoteController, if server supports it.
type MediaServer interface {
	Streamer
	Browser
	RemoteServer
}

// Streamer contains methods for streaming audio from remote location.
type Streamer interface {

	// Stream streams song. If server does not implement separate streaming endpoint,
	// implementcation can wrap Download.
	Stream(Song *models.Song) (io.ReadCloser, interfaces.AudioFormat, error)

	// Download downloads original audio file.
	Download(Song *models.Song) (io.ReadCloser, interfaces.AudioFormat, error)
}

// Browser implements item-based viewing for music artists,albums,playlists etc.
type Browser interface {

	// GetArtists returns all artists
	GetArtists(paging interfaces.Paging) ([]*models.Artist, int, error)

	// GetAlbumArtists returns artists that are marked as album artists. See GetArtists.
	GetAlbumArtists(paging interfaces.Paging) ([]*models.Artist, int, error)
	// GetAlbums gets albums with given paging. Only PageSize and CurrentPage are used. Total count is returned
	GetAlbums(paging interfaces.Paging) ([]*models.Album, int, error)

	// GetArtistAlbums returns albums that artist takes part in.
	GetArtistAlbums(artist models.Id) ([]*models.Album, error)

	// GetAlbumSongs returns songs for given album id.
	GetAlbumSongs(album models.Id) ([]*models.Song, error)
	// GetPlaylists returns all playlists.
	GetPlaylists() ([]*models.Playlist, error)
	// GetPlaylistSongs fills songs array for playlist. If there's error, songs will not be filled
	GetPlaylistSongs(playlist models.Id) ([]*models.Song, error)
	// GetFavoriteArtists returns list of favorite artists.
	GetFavoriteArtists() ([]*models.Artist, error)
	// GetFavoriteAlbums return list of favorite albums.
	GetFavoriteAlbums(paging interfaces.Paging) ([]*models.Album, int, error)

	// GetSimilarArtists returns similar artists for artist id
	GetSimilarArtists(artist models.Id) ([]*models.Artist, error)

	// GetsimilarAlbums returns list of similar albums.
	GetSimilarAlbums(album models.Id) ([]*models.Album, error)

	// GetLatestAlbums returns latest albums.
	GetLatestAlbums() ([]*models.Album, error)

	// GetRecentlyPlayed returns songs that have been played last.
	GetRecentlyPlayed(paging interfaces.Paging) ([]*models.Song, int, error)

	// GetSongs returns songs by paging. It also returns total number of songs.
	GetSongs(page, pageSize int) ([]*models.Song, int, error)

	// GetGenres returns music genres with paging. Return genres, total genres and possible error
	GetGenres(paging interfaces.Paging) ([]*models.IdName, int, error)

	// GetGenreAlbums returns all albums that belong to given genre
	GetGenreAlbums(genre models.IdName) ([]*models.Album, error)

	// GetAlbumArtist returns main artist for album.
	GetAlbumArtist(album *models.Album) (*models.Artist, error)

	// GetInstantMix returns instant mix based on given item.
	GetInstantMix(item models.Item) ([]*models.Song, error)

	// GetLink returns a link to item that can be opened with browser.
	// If there is no link or item is invalid, empty link is returned.
	GetLink(item models.Item) string

	// Search returns values matching query and itemType, limited by number of maxResults,
	// Only items of itemType should ne returned.
	Search(query string, itemType models.ItemType, maxResults int) ([]models.Item, error)

	GetAlbum(id models.Id) (*models.Album, error)

	GetArtist(id models.Id) (*models.Artist, error)

	ImageUrl(item models.Id, itemType models.ItemType) string
}

// RemoteController implents controlling audio player remotely as well as
// keeping remote server updated on player status.
type RemoteController interface {
	// SetPlayer allows connecting remote controller to player, which can
	// then be controlled remotely.
	SetPlayer(player interfaces.Player)

	SetQueue(q interfaces.QueueController)

	RemoteControlEnabled() error
}

// RemoteServer contains general methods for getting server connection status
type RemoteServer interface {
	// GetInfo returns general info
	GetInfo() (*models.ServerInfo, error)

	// ConnectionOk returns nil of connection ok, else returns description for failure.
	ConnectionOk() error

	// GetConfig returns backend config that is saved to config file.
	GetConfig() config.Backend

	// ReportProgress reports player progress to remote controller.
	ReportProgress(state *interfaces.ApiPlaybackState) error

	// Start starts background service for remote server, if any.
	Start() error

	// Stop stops background service for remote server, if any.
	Stop() error
}
