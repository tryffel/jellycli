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
	"tryffel.net/go/jellycli/models"
)

//MusicController gathers all necessary interfaces that can control media and queue plus query item metadata
type MediaController interface {
	QueueController
	PlaybackController
	MediaManager
}

//QueueController controls queue and history
// If no queueChangedCallback is set, no queue updates will be returned
type QueueController interface {
	//GetQueue gets currently ongoing queue of items with complete info for each song
	GetQueue() []*models.SongInfo
	//ClearQueue clears queue. This also calls QueueChangedCallback
	ClearQueue()
	//QueueDuration gets number of queue items
	QueueDuration() int
	//AddItems adds items to the end of queue.
	//Adding items calls QueueChangedCallback
	AddSongs([]*models.Song)
	//Reorder sets item in index currentIndex to newIndex.
	//If either currentIndex or NewIndex is not valid, do nothing.
	//On successful order QueueChangedCallback gets called.
	Reorder(currentIndex, newIndex int)
	//GetHistory get's n past songs that has been played.
	GetHistory(n int) []*models.SongInfo
	//SetQueueChangedCallback sets function that is called every time queue changes.
	SetQueueChangedCallback(func(content []*models.Song))
	//RemoveQueueChangedCallback removes queue changed callback
	RemoveQueueChangedCallback()

	// SetHistoryChangedCallback sets a function that gets called every time history items update
	SetHistoryChangedCallback(func(songs []*models.Song))
}

//PlaybackController controls media playback. Current status is sent to StatusCallback, if set.
type PlaybackController interface {
	//PlayPause toggles pause
	PlayPause()
	//Pause pauses media that's currently playing. If none, do nothing.
	Pause()
	//Continue continues currently paused media.
	Continue()
	//StopMedia stops playing media.
	StopMedia()
	//Next plays currently next item in queue. If there's no next song available, this method does nothing.
	Next()
	//Previous plays last played song (first in history) if there is one.
	Previous()
	//Seek seeks forward given seconds
	Seek(seconds int)
	//SeekBackwards seeks backwards given seconds
	SeekBackwards(seconds int)
	//AddStatusCallback adds callback that get's called every time status has changed,
	//including playback progress
	AddStatusCallback(func(status PlayingState))
	//SetVolume sets volume to given level in range of [0,100]
	SetVolume(level int)
}

//MediaManager manages media: artists, albums, songs
type MediaManager interface {
	SearchArtists(search string) ([]*models.Artist, error)
	SearchAlbums(search string) ([]*models.Album, error)
	GetArtists() ([]*models.Artist, error)
	GetAlbums() ([]*models.Album, error)

	GetArtistAlbums(artist models.Id) ([]*models.Album, error)

	GetAlbumSongs(album models.Id) ([]*models.Song, error)
	GetPlaylists() ([]*models.Album, error)
	GetFavoriteArtists() ([]*models.Artist, error)
	GetFavoriteAlbums() ([]*models.Album, error)

	GetLatestAlbums() ([]*models.Album, error)

	GetStatistics() models.Stats
}

type View int

const (
	ViewAllArtists View = iota
	ViewAllAlbums
	ViewAllSongs
	ViewFavoriteArtists
	ViewFavoriteAlbums
	ViewFavoriteSongs
	ViewPlaylists
	ViewLatestMusic
)
