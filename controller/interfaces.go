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
type ItemController interface {
	//GetChildren returns children for given parent id. If there are none, returns nil
	GetChildren(parent models.Id) []models.Item
	//GetParent returns parent for child id. If there is no parent, return nil
	GetParent(child models.Id) models.Item
	//SetItemsCallback sets callback that gets called when items are retrieved
	SetItemsCallback(func([]models.Id))
	//RemoveItemsCallback removes items callback if there's any
	RemoveItemsCallback()
}

//QueueController controls queue and history
type QueueController interface {
	//GetQueue gets currently ongoing queue of items
	GetQueue() []models.Song
	//ClearQueue clears queue. This also calls QueueChangedCallback
	ClearQueue()
	//QueueDuration gets number of queue items
	QueueDuration() int
	//AddItems adds items to the end of queue.
	//Adding items calls QueueChangedCallback
	AddSongs([]models.Song)
	//Reorder sets item in index currentIndex to newIndex.
	//If either currentIndex or NewIndex is not valid, do nothing.
	//On successful order QueueChangedCallback gets called.
	Reorder(currentIndex, newIndex int)
	//GetHistory get's n past songs that has been played.
	GetHistory(n int) []models.Song
	//SetQueueChangedCallback sets function that is called every time queue changes.
	SetQueueChangedCallback(func(content []models.Song))
	//RemoveQueueChangedCallback removes queue changed callback
	RemoveQueueChangedCallback()
}

type PlaybackController interface {
	Pause()
	Continue()
	Stop()
	Next()
	Previous()
	Seek(seconds int)
	SeekBackwards(seconds int)
	SetStatusCallback(func(state player.PlayingState))
}
