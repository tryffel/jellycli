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

package interfaces

import (
	"math"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
)

//MusicController gathers all necessary interfaces that can control media and queue plus query item metadata
type MediaController interface {
	QueueController
	Player
	ItemController
}

//QueueController controls queue and history
// If no queueChangedCallback is set, no queue updates will be returned
type QueueController interface {
	//GetQueue gets currently ongoing queue of items with complete info for each song
	GetQueue() []*models.Song
	//ClearQueue clears queue. This also calls QueueChangedCallback
	ClearQueue()
	//AddSongs adds songs to the end of queue.
	//Adding songs calls QueueChangedCallback
	AddSongs([]*models.Song)
	//Reorder sets item in index currentIndex to newIndex.
	//If either currentIndex or NewIndex is not valid, do nothing.
	//On successful order QueueChangedCallback gets called.
	Reorder(currentIndex, newIndex int)
	//GetHistory get's n past songs that has been played.
	GetHistory(n int) []*models.Song
	//AddQueueChangedCallback sets function that is called every time queue changes.
	AddQueueChangedCallback(func(content []*models.Song))

	// SetHistoryChangedCallback sets a function that gets called every time history items update
	SetHistoryChangedCallback(func(songs []*models.Song))
}

//MediaManager manages media: artists, albums, songs
type ItemController interface {
	SearchArtists(search string) ([]*models.Artist, error)
	SearchAlbums(search string) ([]*models.Album, error)
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
	GetFavoriteAlbums() ([]*models.Album, error)

	GetLatestAlbums() ([]*models.Album, error)

	// GetStatistics returns application statistics
	GetStatistics() models.Stats

	// GetSongs returns songs by paging. It also returns total number of songs.
	GetSongs(page, pageSize int) ([]*models.Song, int, error)
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
