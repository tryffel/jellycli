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
	"sync"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

// queueItem: song + original index + random index for shuffling
type queueItem struct {
	song *models.Song

	// index is original priority, which is len(queue) at insertion time.
	index int

	// priority is random number between 0-len(queue).
	priority int
}

// queueRbTree implements heap.Interface.
type queueRbTree struct {
	// maxIndex. When adding new item, use maxIndex+1 as item index and increase maxIndex by one.
	maxIndex int

	// is shuffling enabled
	shuffle bool

	tree *rbt.Tree
}

func newQueueHeap() *queueRbTree {
	q := &queueRbTree{
		maxIndex: 0,
		shuffle:  false,
		tree:     rbt.NewWithIntComparator(),
	}
	return q
}

func (q *queueRbTree) Len() int {
	return q.tree.Size()
}

func (q *queueRbTree) Compare(x, y interface{}) int {
	xItem := x.(*queueItem)
	yItem := y.(*queueItem)

	if q.shuffle {
		return xItem.priority - yItem.priority
	} else {
		return xItem.index - yItem.index
	}
}

/* Public methods */
func (q *queueRbTree) SetShuffling(enable bool) {
	if enable == q.shuffle {
		return
	}
	q.shuffle = enable
	items := q.tree.Values()
	q.tree.Clear()
	for _, v := range items {
		queueItem := v.(*queueItem)
		if enable {
			q.tree.Put(queueItem.priority, queueItem)
		} else {
			q.tree.Put(queueItem.index, queueItem)
		}
	}
}

// Clear. First: whether to clear first item too
func (q *queueRbTree) Clear(first bool) {
	if first && q.Len() > 0 {
		q.tree.Clear()
		q.maxIndex = 0
	} else {
		node := q.tree.Left()
		q.tree.Clear()
		if node != nil {
			q.tree.Put(node.Key, node.Value)
			q.maxIndex = node.Key.(int) + 1
		}
	}
}

func (q *queueRbTree) AddSong(song *models.Song, playNext bool, playFirst bool) {
	index := q.maxIndex
	if q.Len() == 1 {
		if playNext {
			node := q.tree.Left()
			if node != nil {
				var key int
				item := node.Value.(*queueItem)
				// we don't care about shuffle
				key = item.index
				q.tree.Remove(key)
				index = item.index
				item.index -= 1
				q.tree.Put(item.index, item)
			}

		} else if playFirst {
			node := q.tree.Left()
			if node != nil {
				item := node.Value.(*queueItem)
				index = item.index - 1
			}
		}
	} else if q.Len() > 1 {
		if playNext {
			node := q.tree.Left()
			if node != nil {
				var key int
				item := node.Value.(*queueItem)
				// we don't care about shuffle yet.
				key = item.index
				q.tree.Remove(key)
				index = item.index

				item.index -= 1
				q.tree.Put(item.index, item)
			}

		} else if playFirst {
			node := q.tree.Left()
			if node != nil {
				item := node.Value.(*queueItem)
				index = item.index - 1
			}
		}
	}

	item := &queueItem{
		song:     song,
		index:    index,
		priority: rand.Int(),
	}
	q.tree.Put(index, item)
	q.maxIndex += 1
}

func (q *queueRbTree) RemoveSong(index int) *models.Song {
	it := q.tree.Iterator()
	i := 0
	for it.Next() {
		if i == index {
			val := it.Value()
			item := val.(*queueItem).song
			key := it.Key()
			q.tree.Remove(key)
			return item
		}
		i += 1
	}
	return nil
}

func (q *queueRbTree) GetQueue() []*models.Song {
	songs := make([]*models.Song, q.Len())
	i := 0
	it := q.tree.Iterator()
	for it.Next() {
		node := it.Value()
		item := node.(*queueItem)
		songs[i] = item.song
		i += 1
	}
	return songs
}

func (q *queueRbTree) GetTotalDuration() interfaces.AudioTick {
	ms := 0
	for it := q.tree.Iterator(); it.Next(); {
		node := it.Value()
		item := node.(*queueItem)
		ms += item.song.Duration
	}
	return interfaces.AudioTick(ms)
}

func (q *queueRbTree) Reorder(index1 int, down bool) {
	index2 := index1 + 1
	if down {
		index2 = index1 - 1
	}

	it := q.tree.Iterator()
	i := 0
	var item1 *queueItem
	var item2 *queueItem
	for it.Next() {

		if i == index1 {
			node := it.Value()
			item1 = node.(*queueItem)
		}
		if i == index2 {
			node := it.Value()
			item2 = node.(*queueItem)
		}
		if item1 != nil && item2 != nil {
			break
		}
		i += 1
	}
	if item1 == nil || item2 == nil {
		logrus.Errorf("did not find both elements from queue to re-order")
	} else {
		item1.song, item2.song = item2.song, item1.song
	}
}

// Queue implements interfaces.QueueController
type Queue struct {
	lock               sync.RWMutex
	tree               *queueRbTree
	history            []*models.Song
	queueUpdatedFunc   []func([]*models.Song)
	historyUpdatedFunc func([]*models.Song)
}

func newQueue() *Queue {
	q := &Queue{
		tree:             newQueueHeap(),
		history:          []*models.Song{},
		queueUpdatedFunc: make([]func([]*models.Song), 0),
	}
	return q
}

// GetQueue gets currently ongoing queue of items with complete info for each song.
func (q *Queue) GetQueue() []*models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.tree.GetQueue()
}

// ClearQueue clears queue. This also calls QueueChangedCallback.
func (q *Queue) ClearQueue(first bool) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyQueueUpdated()
	q.tree.Clear(first)
}

// AddSongs adds songs to the end of queue.
// Adding songs calls QueueChangedCallback.
func (q *Queue) AddSongs(songs []*models.Song) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.notifyQueueUpdated()

	for _, v := range songs {
		q.tree.AddSong(v, false, false)
	}

	logrus.Debug("Adding songs to queue, current size: ", q.tree.Len())
}

func (q *Queue) PlayNext(songs []*models.Song) {
	q.lock.Lock()
	for i := len(songs); i > 0; i-- {
		q.tree.AddSong(songs[i-1], true, false)
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
		q.tree.RemoveSong(index)
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

	heapLen := q.tree.Len()

	//var item *models.Song
	if heapLen == 0 {
		// no action
	} else if index < 0 || index > heapLen-1 {
		// illegal index
	} else if index >= 0 && index < heapLen-2 && !down {
		changed = true
		q.tree.Reorder(index, down)
	} else if index >= 1 && down {
		changed = true
		q.tree.Reorder(index, down)
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

	songs := q.tree.GetQueue()
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
	if q.tree.Len() == 0 {
		return
	}

	song := q.tree.RemoveSong(0)
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
	q.tree.AddSong(song, false, true)
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
	return q.tree.Len() == 0
}

func (q *Queue) currentSong() *models.Song {
	q.lock.RLock()
	defer q.lock.RUnlock()
	//if len(q.items) > 0 {
	//	return q.items[0]
	//} else {
	//	return &models.Song{}
	//}
	return nil
}

func (q *Queue) SetShuffle(enabled bool) {
	q.lock.Lock()
	changed := enabled != q.tree.shuffle
	if !changed {
		q.lock.Unlock()
		return
	}
	q.tree.SetShuffling(enabled)
	q.lock.Unlock()
	q.notifyQueueUpdated()
}

func (q *Queue) GetShuffle() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.tree.shuffle
}
