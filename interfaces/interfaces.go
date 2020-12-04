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

// Package interfaces contains interfaces and structs that multiple packages use and communicate with.
package interfaces

import (
	"math"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
)

// QueueController controls queue and history. Queue shows only upcoming songs and first item in queue is being
// currently played. When moving to next item in queue, first item is moved to history.
// If no queueChangedCallback is set, no queue updates will be returned
type QueueController interface {
	//GetQueue gets currently ongoing queue of items with complete info for each song
	GetQueue() []*models.Song
	//ClearQueue clears queue. This also calls QueueChangedCallback. If first = true, clear also first item. Else
	// leave it as it is.
	ClearQueue(first bool)
	//AddSongs adds songs to the end of queue.
	//Adding songs calls QueueChangedCallback
	AddSongs([]*models.Song)

	//PlayNext adds songs to 2nd index in order.
	PlayNext([]*models.Song)
	//Reorder sets item in index currentIndex to newIndex.
	//If either currentIndex or NewIndex is not valid, do nothing.
	//On successful order QueueChangedCallback gets called.

	// Reorder shifts item in current index to left or right (earlier / later) by one depending on left.
	// If down, play it earlier, else play it later. Returns true if reorder was made.
	Reorder(currentIndex int, down bool) bool
	//GetHistory get's n past songs that has been played.
	GetHistory(n int) []*models.Song
	//AddQueueChangedCallback sets function that is called every time queue changes.
	AddQueueChangedCallback(func(content []*models.Song))

	// RemoveSongs remove song in given index. First index is 0.
	RemoveSong(index int)

	// SetHistoryChangedCallback sets a function that gets called every time history items update
	SetHistoryChangedCallback(func(songs []*models.Song))
}

//MediaManager manages media: artists, albums, songs
type ItemController interface {
	// Search returns list of items based on search query. Item types
	// Queue and history returns error.
	Search(itemType models.ItemType, query string) ([]models.Item, error)
	// GetArtists gets artist with given paging. Only PageSize and CurrentPage are used. Total count is returned
	GetArtists(paging Paging) ([]*models.Artist, int, error)

	// GetAlbumArtists returns artists that are marked as album artists. See GetArtists.
	GetAlbumArtists(paging Paging) ([]*models.Artist, int, error)
	// GetAlbums gets albums with given paging. Only PageSize and CurrentPage are used. Total count is returned
	GetAlbums(paging Paging) ([]*models.Album, int, error)

	GetArtistAlbums(artist models.Id) ([]*models.Album, error)

	GetAlbumSongs(album models.Id) ([]*models.Song, error)
	GetPlaylists() ([]*models.Playlist, error)
	// GetPlaylistSongs fills songs array for playlist. If there's error, songs will not be filled
	GetPlaylistSongs(playlist *models.Playlist) error
	GetFavoriteArtists() ([]*models.Artist, error)
	GetFavoriteAlbums(paging Paging) ([]*models.Album, int, error)

	// GetSimilarArtists returns similar artists for artist id
	GetSimilarArtists(artist models.Id) ([]*models.Artist, error)

	GetSimilarAlbums(album models.Id) ([]*models.Album, error)

	GetLatestAlbums() ([]*models.Album, error)

	GetRecentlyPlayed(paging Paging) ([]*models.Song, int, error)

	// GetStatistics returns application statistics
	GetStatistics() models.Stats

	// GetSongs returns songs by paging. It also returns total number of songs.
	GetSongs(page, pageSize int) ([]*models.Song, int, error)

	// GetGenres returns music genres with paging. Return genres, total genres and possible error
	GetGenres(paging Paging) ([]*models.IdName, int, error)

	// GetGenreAlbums returns all albums that belong to given genre
	GetGenreAlbums(genre models.IdName) ([]*models.Album, error)

	GetAlbumArtist(album *models.Album) (*models.Artist, error)

	GetSongArtistAlbum(song *models.Song) (*models.Album, *models.Artist, error)

	// GetInstantMix returns instant mix based on given item.
	GetInstantMix(item models.Item) ([]*models.Song, error)

	// GetLink returns a link to item that can be opened with browser.
	// If there is no link or item is invalid, empty link is returned.
	GetLink(item models.Item) string
}

// Paging. First page is 0
type Paging struct {
	TotalItems  int
	TotalPages  int
	CurrentPage int
	PageSize    int
}

// DefaultPaging returns paging with page 0 and default pagesize
func DefaultPaging() Paging {
	return Paging{
		TotalItems:  0,
		TotalPages:  0,
		CurrentPage: 0,
		PageSize:    config.PageSize,
	}
}

// SetTotalItems calculates number of pages for current page size
func (p *Paging) SetTotalItems(count int) {
	p.TotalItems = count
	p.TotalPages = int(math.Ceil(float64(count) / float64(p.PageSize)))
}

// Offset returns offset
func (p *Paging) Offset() int {
	return p.PageSize * p.CurrentPage
}
