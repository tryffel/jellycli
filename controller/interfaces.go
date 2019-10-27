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
	"tryffel.net/pkg/jellycli/models"
	"tryffel.net/pkg/jellycli/player"
)

//MusicController gathers all necessary interfaces that can control media and queue plus query item metadata
type MediaController interface {
	ItemController
	QueueController
	PlaybackController
}

//ItemController retrieves children and returns them with ItemsCallback
//Manages item metadata and not the files themselves.
// If itemsCallback is not set, no results will be retrieved.
type ItemController interface {
	//GetItem gets item with given id. If none found or if errors, return nil
	GetItem(id models.Id, itemType models.ItemType)

	//GetItems get multiple items for given ids.
	GetItems(ids []models.Id, itemType models.ItemType)

	//GetChildren returns children for given parent id. If there are none, returns nil
	GetChildren(parent models.Id, parentType models.ItemType)
	//GetParent returns parent for child id. If there is no parent, return nil
	GetParent(child models.Id, childType models.ItemType)
	//SetItemsCallback sets callback that gets called when items are retrieved
	SetItemsCallback(func([]models.Item))
	//RemoveItemsCallback removes items callback if there's any
	RemoveItemsCallback()
	//GetDefaultItems()
}

//QueueController controls queue and history
// If no queueChangedCallback is set, no queue updates will be returned
type QueueController interface {
	//GetQueue gets currently ongoing queue of items
	GetQueue() []*models.Song
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
	GetHistory(n int) []*models.Song
	//SetQueueChangedCallback sets function that is called every time queue changes.
	SetQueueChangedCallback(func(content []*models.Song))
	//RemoveQueueChangedCallback removes queue changed callback
	RemoveQueueChangedCallback()
}

//PlaybackController controls media playback. Current status is sent to StatusCallback, if set.
type PlaybackController interface {
	//Pause pauses media that's currently playing. If none, do nothing.
	Pause()
	//Continue continues currently paused media.
	Continue()
	//StopMedia stops playing media.
	StopMedia()
	//Next plays currently next item in queue, if any.
	Next()
	//Previous plays last played song (first in history)
	Previous()
	//Seek seeks forward given seconds
	Seek(seconds int)
	//SeekBackwards seeks backwards given seconds
	SeekBackwards(seconds int)
	//SetStatusCallback sets callback that get's called every time status has changed,
	//including playback progress
	SetStatusCallback(func(state player.PlayingState))
	//SetVolume sets volume to given level in range of [0,100]
	SetVolume(level int)
}
