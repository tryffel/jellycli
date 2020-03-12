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

package player

import (
	"github.com/sirupsen/logrus"
	"sync"
	"tryffel.net/go/jellycli/models"
)

// Queue implements interfaces.QueueController
type Queue struct {
	lock               sync.RWMutex
	items              []*models.Song
	history            []*models.Song
	queueUpdatedFunc   []func([]*models.Song)
	historyUpdatedFunc func([]*models.Song)
}

func newQueue() *Queue {
	q := &Queue{
		items:            []*models.Song{},
		history:          []*models.Song{},
		queueUpdatedFunc: make([]func([]*models.Song), 0),
	}
	return q
}

// GetQueue gets currently ongoing queue of items with complete info for each song.
func (q *Queue) GetQueue() []*models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.items
}

// ClearQueue clears queue. This also calls QueueChangedCallback.
func (q *Queue) ClearQueue() {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyQueueUpdated()
	q.items = []*models.Song{}
}

// AddSongs adds songs to the end of queue.
// Adding songs calls QueueChangedCallback.
func (q *Queue) AddSongs(songs []*models.Song) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyQueueUpdated()
	q.items = append(q.items, songs...)
	logrus.Debug("Adding songs to queue, current size: ", len(q.items))
}

// Reorder sets item in index currentIndex to newIndex.
// If either currentIndex or NewIndex is not valid, do nothing.
// On successful order QueueChangedCallback gets called.
func (q *Queue) Reorder(currentIndex, newIndex int) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyQueueUpdated()
	//TODO: Fix ordering songs
	if currentIndex < 0 || newIndex < 0 {
		return
	}
	last := len(q.items) - 1
	if currentIndex > last || newIndex > last {
		return
	}

	song := q.items[currentIndex]

	var temp []*models.Song
	if currentIndex < newIndex {
		temp = append(q.items[:currentIndex], q.items[currentIndex+1:newIndex+1]...)
		temp = append(temp, song)
		temp = append(temp, q.items[newIndex+1:]...)
	} else {
		if newIndex == 0 {
			temp = []*models.Song{song}
		} else {
			temp = append(q.items[:newIndex], song)
		}
		temp = append(temp, q.items[newIndex:currentIndex]...)
	}
	q.items = temp
}

// GetHistory get's n past songs that has been played.
func (q *Queue) GetHistory(n int) []*models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if n > len(q.history) {
		return q.history
	}
	return q.history[:n]
}

// AddQueueChangedCallback adds function that is called every time queue changes.
func (q *Queue) AddQueueChangedCallback(cb func(content []*models.Song)) {
	q.queueUpdatedFunc = append(q.queueUpdatedFunc, cb)
}

// SetHistoryChangedCallback sets a function that gets called every time history items update.
func (q *Queue) SetHistoryChangedCallback(cb func([]*models.Song)) {
	q.historyUpdatedFunc = cb
}

func (q *Queue) notifyQueueUpdated() {
	if q.queueUpdatedFunc == nil {
		return
	}
	for _, v := range q.queueUpdatedFunc {
		v(q.items)
	}
}

func (q *Queue) notifyHistoryUpdated() {
	if q.historyUpdatedFunc != nil {
		q.historyUpdatedFunc(q.history)
	}
}

// remove first song from queue and move to history
func (q *Queue) songComplete() {
	q.lock.Lock()
	defer q.notifyQueueUpdated()
	defer q.notifyHistoryUpdated()
	if len(q.items) == 0 {
		return
	}
	song := q.items[0]
	q.items = q.items[1:]
	if q.history == nil {
		q.history = []*models.Song{song}
	} else {
		q.history = append([]*models.Song{song}, q.history...)
	}
	q.lock.Unlock()
}

// remove first item from history and move to queue
func (q *Queue) playLastSong() {
	q.lock.Lock()
	defer q.notifyQueueUpdated()
	defer q.notifyHistoryUpdated()
	if len(q.history) == 0 {
		return
	}
	song := q.history[0]
	q.items = append([]*models.Song{song}, q.items...)
	if q.history == nil {
		q.history = q.history[1:]
	} else {
		q.history = q.history[1:]
	}
	q.lock.Unlock()
}

func (q *Queue) empty() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.items) == 0
}

func (q *Queue) currentSong() *models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if len(q.items) > 0 {
		return q.items[0]
	} else {
		return &models.Song{}
	}
}
