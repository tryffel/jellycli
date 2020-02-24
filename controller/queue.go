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
	"github.com/sirupsen/logrus"
	"sync"
	"tryffel.net/go/jellycli/models"
)

type queue struct {
	lock     sync.RWMutex
	items    []*models.Song
	history  []*models.Song
	updateCb func([]*models.Song)
}

func newQueue() *queue {
	q := &queue{
		items:    []*models.Song{},
		history:  []*models.Song{},
		updateCb: nil,
	}
	return q
}

func (q *queue) GetQueue() []*models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.items
}

func (q *queue) ClearQueue() {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyUpdates()
	q.items = []*models.Song{}
}

func (q *queue) QueueDuration() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	duration := 0
	for _, v := range q.items {
		duration += v.Duration
	}
	return duration
}

func (q *queue) AddSongs(songs []*models.Song) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyUpdates()
	q.items = append(q.items, songs...)
	logrus.Debug("Adding songs to queue, current size: ", len(q.items))
}

func (q *queue) Reorder(currentIndex, newIndex int) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyUpdates()
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

func (q *queue) GetHistory(n int) []*models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if n > len(q.history) {
		return q.history
	}
	return q.history[:n]
}

func (q *queue) SetQueueChangedCallback(cb func(content []*models.Song)) {
	q.updateCb = cb
}

func (q *queue) RemoveQueueChangedCallback() {
	q.updateCb = nil
}

func (q *queue) notifyUpdates() {
	if q.updateCb == nil {
		return
	}
	q.updateCb(q.items)
}

func (q *queue) songComplete() {
	q.lock.Lock()
	defer q.lock.Unlock()
	logrus.Debugf("Song (%s) complete, remove from queue", q.items[0].Name)
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
}

func (q *queue) empty() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.items) == 0
}

func (q *queue) currentSong() *models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if len(q.items) > 0 {
		return q.items[0]
	} else {
		return &models.Song{}
	}
}
