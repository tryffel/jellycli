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
	"math/rand"
	"sort"
	"sync"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

// queueItem: song + original index + random index for shuffling
type queueItem struct {
	song *models.Song

	// index is original priority, which is len(queue) at insertion time.
	index int

	// priority is random number between 0-len(queue).
	priority int
}

// queueList implements sort.Interface.
type queueList struct {
	items []*queueItem

	// maxIndex. When adding new item, use maxIndex+1 as item index and increase maxIndex by one.
	maxIndex int

	// is shuffling enabled
	shuffle bool
}

func (q *queueList) Less(i, j int) bool {
	if q.shuffle {
		return q.items[i].priority < q.items[j].priority
	} else {
		return q.items[i].index < q.items[j].index
	}
}

func (q *queueList) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func newQueueList() *queueList {
	q := &queueList{
		maxIndex: 0,
		shuffle:  false,
		items:    make([]*queueItem, 0),
	}
	return q
}

func (q *queueList) Len() int {
	return len(q.items)
}

func (q *queueList) SetShuffling(enable bool) {
	if enable == q.shuffle {
		return
	}
	q.shuffle = enable

	if enable && len(q.items) > 0 {
		// make sure 1st stays 1st after shuffling
		q.items[0].priority = 0
		for _, v := range q.items[1:] {
			v.priority = rand.Int()
		}
	}
	sort.Sort(q)
}

// Clear. First: whether to clear first item too
func (q *queueList) Clear(first bool) {
	if q.Len() == 0 {
		return
	}
	if first {
		q.items = make([]*queueItem, 0)
		q.maxIndex = 0
	} else {
		q.items = []*queueItem{q.items[0]}
	}
}

func (q *queueList) AddSong(song *models.Song, playNext bool, playFirst bool) {
	index := q.maxIndex
	priority := rand.Int()
	needsSort := false

	if len(q.items) == 0 {
	} else if q.shuffle && playFirst {
		priority = q.items[0].priority - 1
		needsSort = true
	} else if playFirst {
		index = q.items[0].index - 1
	} else if playNext {
		index = q.items[0].index
		q.items[0].index -= 1
	} else {
		// normal insertion
	}

	item := &queueItem{
		song:     song,
		index:    index,
		priority: priority,
	}

	if len(q.items) == 0 || q.shuffle {
		q.items = append(q.items, item)
	} else if playNext {
		temp := append([]*queueItem{q.items[0]}, item)
		q.items = append(temp, q.items[1:]...)
	} else if playFirst {
		q.items = append([]*queueItem{item}, q.items...)

	} else {
		q.items = append(q.items, item)
	}
	q.maxIndex += 1
	if needsSort {
		sort.Sort(q)
	}
}

func (q *queueList) RemoveSong(index int) (song *models.Song) {
	if q.Len() == 0 {
		return nil
	}

	if q.Len() == 1 && index == 1 {
		song = q.items[0].song
		q.Clear(true)
		return
	}

	if len(q.items) > index+1 {
		song = q.items[index].song
		q.items = append(q.items[:index], q.items[index+1:]...)
	} else if len(q.items) == index+1 {
		song = q.items[index].song
		q.items = q.items[:index]
	}
	return
}

func (q *queueList) GetQueue() []*models.Song {
	songs := make([]*models.Song, q.Len())
	for i, v := range q.items {
		songs[i] = v.song
	}
	return songs
}

func (q *queueList) GetTotalDuration() interfaces.AudioTick {
	ms := 0
	for _, v := range q.items {
		ms += v.song.Duration
	}
	return interfaces.AudioTick(ms)
}

func (q *queueList) Reorder(index1 int, down bool) {
	index2 := index1 + 1
	if down {
		index2 = index1 - 1
	}

	// no point reordering shuffled songs
	if q.shuffle {
		return
	}

	if index2 < len(q.items) && index1 < len(q.items) {
		q.items[index1].index, q.items[index2].index = q.items[index2].index, q.items[index1].index
		sort.Sort(q)
	}
}

// Queue implements interfaces.QueueController
type Queue struct {
	lock               sync.RWMutex
	list               *queueList
	history            []*models.Song
	queueUpdatedFunc   []func([]*models.Song)
	historyUpdatedFunc func([]*models.Song)
}

func newQueue() *Queue {
	q := &Queue{
		list:             newQueueList(),
		history:          []*models.Song{},
		queueUpdatedFunc: make([]func([]*models.Song), 0),
	}
	return q
}

// GetQueue gets currently ongoing queue of items with complete info for each song.
func (q *Queue) GetQueue() []*models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.list.GetQueue()
}

// ClearQueue clears queue. This also calls QueueChangedCallback.
func (q *Queue) ClearQueue(first bool) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyQueueUpdated()
	q.list.Clear(first)
}

// AddSongs adds songs to the end of queue.
// Adding songs calls QueueChangedCallback.
func (q *Queue) AddSongs(songs []*models.Song) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyQueueUpdated()

	for _, v := range songs {
		q.list.AddSong(v, false, false)
	}

	logrus.Debug("Adding songs to queue, current size: ", q.list.Len())
}

func (q *Queue) PlayNext(songs []*models.Song) {
	q.lock.Lock()
	for i := len(songs); i > 0; i-- {
		q.list.AddSong(songs[i-1], true, false)
	}
	q.lock.Unlock()
	q.notifyQueueUpdated()
}

func (q *Queue) RemoveSong(index int) {
	changed := false
	q.lock.Lock()
	if index == 0 {
		// if we remove first, we must notify player to move to next song
	} else {
		q.list.RemoveSong(index)
		changed = true
	}
	q.lock.Unlock()
	if changed {
		q.notifyQueueUpdated()
	}
}

// Reorder sets item in index currentIndex to newIndex.
// If either currentIndex or NewIndex is not valid, do nothing.
// On successful order QueueChangedCallback gets called.
func (q *Queue) Reorder(index int, down bool) bool {
	q.lock.Lock()
	changed := false

	heapLen := q.list.Len()

	//var item *models.Song
	if heapLen == 0 {
		// no action
	} else if index < 0 || index > heapLen-1 {
		// illegal index
	} else if index >= 0 && index < heapLen-2 && !down {
		changed = true
		q.list.Reorder(index, down)
	} else if index >= 1 && down {
		changed = true
		q.list.Reorder(index, down)
	}

	q.lock.Unlock()
	if changed {
		q.notifyQueueUpdated()
	}

	return changed
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

	songs := q.list.GetQueue()
	for _, v := range q.queueUpdatedFunc {
		v(songs)
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
	if q.list.Len() == 0 {
		return
	}

	song := q.list.RemoveSong(0)
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
		q.lock.Unlock()
		return
	}
	song := q.history[0]
	q.list.AddSong(song, false, true)
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
	return q.list.Len() == 0
}

func (q *Queue) SetShuffle(enabled bool) {
	q.lock.Lock()
	changed := enabled != q.list.shuffle
	if !changed {
		q.lock.Unlock()
		return
	}
	q.list.SetShuffling(enabled)
	q.lock.Unlock()
	q.notifyQueueUpdated()
}
