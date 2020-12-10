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
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
	"tryffel.net/go/jellycli/models"
)

func testSongs() []*models.Song {
	return []*models.Song{
		{Id: "song-1", Name: "song-1", Duration: 60},
		{Id: "song-2", Name: "song-2", Duration: 10},
		{Id: "song-3", Name: "song-3", Duration: 1},
		{Id: "song-4", Name: "song-4", Duration: 350},
		{Id: "song-5", Name: "song-5", Duration: 10},
		{Id: "song-6", Name: "song-6", Duration: 10},
		{Id: "song-7", Name: "song-7", Duration: 10},
		{Id: "song-8", Name: "song-8", Duration: 80},
		{Id: "song-9", Name: "song-9", Duration: 80},
	}
}

func Test_queue_GetQueue(t *testing.T) {
	tests := []struct {
		name  string
		songs []*models.Song
	}{
		{
			songs: []*models.Song{},
		},
		{
			songs: []*models.Song{
				{
					Id:   "song-a",
					Name: "song-a",
				},
				{
					Id:   "song-b",
					Name: "song-b",
				},
				{
					Id:   "song-c",
					Name: "song-c",
				},
			},
		},
		{
			songs: testSongs(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newQueue()
			q.AddSongs(tt.songs)
			if got := q.GetQueue(); !reflect.DeepEqual(got, tt.songs) {
				t.Errorf("GetQueue() = %v, want %v", got, tt.songs)
			}
		})
	}
}

func TestQueue_RemoveSong(t *testing.T) {
	queue := newQueue()
	songs := testSongs()

	tests := []struct {
		name      string
		songs     []*models.Song
		index     int
		wantSongs []*models.Song
	}{
		{
			songs:     []*models.Song{},
			index:     0,
			wantSongs: []*models.Song{},
		},
		{
			songs:     songs,
			index:     len(songs) - 1,
			wantSongs: songs[:(len(songs) - 1)],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue.ClearQueue(true)
			queue.AddSongs(tt.songs)

			queue.RemoveSong(tt.index)

			gotSongs := queue.GetQueue()
			logDiff(t, tt.wantSongs, gotSongs, "songs after removal")
		})
	}
}

func Test_queue_Reorder(t *testing.T) {
	songs := testSongs()

	type ordering struct {
		from int
		down bool
	}

	tests := []struct {
		name      string
		songs     []*models.Song
		want      []*models.Song
		orderings []ordering
		shuffle   bool
	}{
		{
			name:  "first-to-right",
			songs: songs,
			want: []*models.Song{
				songs[1], songs[0], songs[2], songs[3], songs[4], songs[5], songs[6],
				songs[7], songs[8],
			},
			orderings: []ordering{
				{0, false},
			},
			shuffle: false,
		},
		{
			// no edit
			name:  "first-to-left",
			songs: songs,
			want:  songs,
			orderings: []ordering{
				{0, true},
			},
			shuffle: false,
		},
		{
			name:  "2nd-to-3rd",
			songs: songs,
			want: []*models.Song{
				songs[0], songs[2], songs[1], songs[3], songs[4], songs[5], songs[6],
				songs[7], songs[8],
			},
			orderings: []ordering{
				{1, false},
			},
			shuffle: false,
		},
		{
			name:  "4nd-to-3rd",
			songs: songs,
			want: []*models.Song{
				songs[0], songs[1], songs[3], songs[2], songs[4], songs[5], songs[6],
				songs[7], songs[8],
			},
			orderings: []ordering{
				{3, true},
			},
			shuffle: false,
		},
		{
			name:  "last-left",
			songs: songs,
			want: []*models.Song{
				songs[0], songs[1], songs[2], songs[3], songs[4], songs[5], songs[6],
				songs[8], songs[7],
			},
			orderings: []ordering{
				{8, true},
			},
			shuffle: false,
		},
		{
			name:  "last-right",
			songs: songs,
			want: []*models.Song{
				songs[0], songs[1], songs[2], songs[3], songs[4], songs[5], songs[6],
				songs[7], songs[8],
			},
			orderings: []ordering{
				{8, false},
			},
			shuffle: false,
		},
		{
			name:  "negative",
			songs: songs,
			want: []*models.Song{
				songs[0], songs[1], songs[2], songs[3], songs[4], songs[5], songs[6],
				songs[7], songs[8],
			},
			orderings: []ordering{
				{-1, false},
			},
			shuffle: false,
		},
		{
			name:  "Reorder while shuffle",
			songs: songs,
			want:  songs,
			orderings: []ordering{
				{8, false},
			},
			shuffle: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newQueue()
			q.AddSongs(songs)
			q.SetShuffle(tt.shuffle)
			for _, v := range tt.orderings {
				q.Reorder(v.from, v.down)
			}

			if tt.shuffle {
				q.SetShuffle(false)
			}
			got := q.GetQueue()

			logDiff(t, got, tt.want, "reorder")
		})
	}
}

func Test_queue_songComplete(t *testing.T) {
	songs := testSongs()
	tests := []struct {
		name     string
		songs    []*models.Song
		complete int
		want     []*models.Song
	}{
		{
			songs:    songs,
			complete: 1,
			want:     []*models.Song{songs[0]},
		},
		{
			songs:    songs,
			complete: 4,
			want:     []*models.Song{songs[3], songs[2], songs[1], songs[0]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newQueue()
			q.AddSongs(tt.songs)
			for i := 0; i < tt.complete; i++ {
				q.songComplete()
			}

			history := q.GetHistory(10)
			if !reflect.DeepEqual(history, tt.want) {
				t.Errorf("TestQueue songComplete history: got %v, want: %v", history, tt.want)
			}

			songs := q.GetQueue()
			wantSongs := tt.songs[tt.complete:]
			diff := cmp.Diff(songs, wantSongs)
			if diff != "" {
				t.Errorf("TestQueue songComplete remove items: %s", diff)
			}
		})
	}
}

func Test_queue_AddSongs(t *testing.T) {
	songs := testSongs()
	tests := []struct {
		songs   []*models.Song
		name    string
		add     []*models.Song
		want    []*models.Song
		shuffle bool
	}{
		{
			songs:   songs,
			add:     []*models.Song{songs[1], songs[2], songs[3]},
			want:    append(songs, songs[1], songs[2], songs[3]),
			shuffle: false,
		},
		{
			songs:   nil,
			add:     []*models.Song{songs[1], songs[2], songs[3]},
			want:    []*models.Song{songs[1], songs[2], songs[3]},
			shuffle: false,
		},
		{
			songs:   songs,
			add:     nil,
			want:    songs,
			shuffle: false,
		},
		{
			name:    "add songs while in shuffle mode",
			songs:   nil,
			add:     []*models.Song{songs[1], songs[2], songs[3]},
			want:    []*models.Song{songs[1], songs[2], songs[3]},
			shuffle: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newQueue()
			q.SetShuffle(tt.shuffle)
			q.AddSongs(tt.songs)
			q.AddSongs(tt.add)

			if tt.shuffle {
				q.SetShuffle(false)
			}
			got := q.GetQueue()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddSongs, got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestQueue_playLastSong(t *testing.T) {
	songs := testSongs()
	tests := []struct {
		name        string
		songs       []*models.Song
		queue       []*models.Song
		history     []*models.Song
		wantQueue   []*models.Song
		wantHistory []*models.Song
		// how many songs
		previous int
		shuffle  bool
	}{
		{
			// simple case
			songs:       songs,
			previous:    1,
			queue:       []*models.Song{songs[0]},
			history:     []*models.Song{songs[1]},
			wantQueue:   []*models.Song{songs[1], songs[0]},
			wantHistory: []*models.Song{},
		},
		{
			// more rounds
			songs:       songs,
			previous:    4,
			queue:       []*models.Song{songs[0], songs[1]},
			history:     []*models.Song{songs[1], songs[2], songs[3], songs[4]},
			wantQueue:   []*models.Song{songs[4], songs[3], songs[2], songs[1], songs[0], songs[1]},
			wantHistory: []*models.Song{},
		},
		{
			// not enough songs to play from
			songs:       songs,
			previous:    3,
			queue:       []*models.Song{songs[0], songs[1]},
			history:     []*models.Song{songs[1]},
			wantQueue:   []*models.Song{songs[1], songs[0], songs[1]},
			wantHistory: []*models.Song{},
		},
		{
			name:        "shuffle",
			songs:       songs,
			previous:    1,
			queue:       []*models.Song{songs[0]},
			history:     []*models.Song{songs[1]},
			wantQueue:   []*models.Song{songs[1], songs[0]},
			wantHistory: []*models.Song{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newQueue()
			q.history = tt.history
			q.SetShuffle(tt.shuffle)
			q.AddSongs(tt.queue)
			for i := 0; i < tt.previous; i++ {
				q.playLastSong()
			}

			q.SetShuffle(false)

			history := q.GetHistory(10)
			queue := q.GetQueue()
			if !reflect.DeepEqual(history, tt.wantHistory) {
				t.Errorf("TestQueue playLastSong history: got %v, want: %v", history, tt.wantHistory)
			}

			if !reflect.DeepEqual(queue, tt.wantQueue) {
				t.Errorf("TestQueue playLastSong queue: got %v, want: %v",
					queue, tt.wantQueue)
			}
		})
	}
}

func TestQueue_PlayNext(t *testing.T) {
	songs := testSongs()
	type fields struct {
		items              []*models.Song
		history            []*models.Song
		queueUpdatedFunc   []func([]*models.Song)
		historyUpdatedFunc func([]*models.Song)
	}
	type args struct {
		songs []*models.Song
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantQueue []*models.Song
	}{
		{
			fields: fields{
				items:              []*models.Song{},
				history:            []*models.Song{},
				queueUpdatedFunc:   nil,
				historyUpdatedFunc: nil,
			},
			args:      args{songs: []*models.Song{songs[0]}},
			wantQueue: []*models.Song{songs[0]},
		},
		{
			fields: fields{
				items:              []*models.Song{songs[0]},
				history:            []*models.Song{},
				queueUpdatedFunc:   nil,
				historyUpdatedFunc: nil,
			},
			args:      args{songs: []*models.Song{songs[1]}},
			wantQueue: []*models.Song{songs[0], songs[1]},
		},
		{
			fields: fields{
				items:              []*models.Song{songs[0], songs[1], songs[2]},
				history:            []*models.Song{},
				queueUpdatedFunc:   nil,
				historyUpdatedFunc: nil,
			},
			args:      args{songs: []*models.Song{songs[4], songs[5]}},
			wantQueue: []*models.Song{songs[0], songs[4], songs[5], songs[1], songs[2]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				list:               newQueueList(),
				history:            tt.fields.history,
				queueUpdatedFunc:   tt.fields.queueUpdatedFunc,
				historyUpdatedFunc: tt.fields.historyUpdatedFunc,
			}
			q.AddSongs(tt.fields.items)
			q.PlayNext(tt.args.songs)
			if !reflect.DeepEqual(q.GetQueue(), tt.wantQueue) {
				t.Errorf("queue playNext, want: %v, got: %v", tt.wantQueue, q.GetQueue())
			}
		})
	}
}

func TestQueue_Shuffle(t *testing.T) {

	songs := testSongs()
	q := newQueue()

	tests := []struct {
		name  string
		songs []*models.Song
	}{
		{
			name:  "simple",
			songs: songs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q.ClearQueue(true)
			q.SetShuffle(false)
			q.AddSongs(songs)

			originalSongs := q.GetQueue()
			if !reflect.DeepEqual(originalSongs, tt.songs) {
				t.Errorf("No shuffle = %v, want %v", originalSongs, tt.songs)
			}

			q.SetShuffle(true)
			shuffleCollection := make(map[string]bool, len(songs))
			shuffledSongs := q.GetQueue()

			if len(shuffledSongs) != len(songs) {
				t.Errorf("shuffled songs len differs from original")
			}

			if reflect.DeepEqual(shuffledSongs, tt.songs) {
				// guess sometimes shuffle matches original, try again once and hope it does not match twice.
				q.SetShuffle(false)
				q.SetShuffle(true)
				if reflect.DeepEqual(originalSongs, tt.songs) {
					t.Errorf("Shuffle returned original songs")
				}
			}

			for _, v := range shuffledSongs {
				if shuffleCollection[v.Id.String()] {
					t.Errorf("duplicate song in shuffled array: %v", v.Id)
				}
				shuffleCollection[v.Id.String()] = true
			}

			if shuffledSongs[0] != originalSongs[0] {
				t.Errorf("1st shuffled song does not match original song")
			}

			q.SetShuffle(false)
			if !reflect.DeepEqual(originalSongs, tt.songs) {
				t.Errorf("undo shuffle = %v, want %v", originalSongs, tt.songs)
			}
		},
		)
	}
}

func TestQueue_Complete(t *testing.T) {

	songs := testSongs()
	q := newQueue()

	empty := []*models.Song{}

	q.AddSongs(songs)
	logDiff(t, songs, q.GetQueue(), "initial queue")
	q.ClearQueue(true)
	logDiff(t, empty, q.GetQueue(), "clear queue completely")

	q.AddSongs(songs)
	q.ClearQueue(false)
	logDiff(t, []*models.Song{songs[0]}, q.GetQueue(), "clear queue, leave first")

	q.AddSongs(songs)
	logDiff(t, append([]*models.Song{songs[0]}, songs...), q.GetQueue(), "clear queue, leave first")

	q.SetShuffle(true)
	shuffledQueue := q.GetQueue()
	q.songComplete()
	logDiff(t, shuffledQueue[1:], q.GetQueue(), "song complete during shuffle")

	q.SetShuffle(false)
	queue := q.GetQueue()
	// we have added 1 and removed 1 song, lists should be equal in length
	logDiff(t, len(songs), len(queue), "undo shuffling, test queue size decreased")
}

func TestQueue_TestAddSongsDuringShouffle(t *testing.T) {
	// 1. add songs
	// 2. shuffle songs
	// 3. add couple of songs
	// 4. remove 1st song
	// 5. undo shuffling

	q := newQueue()
	songs := testSongs()
	q.AddSongs(songs)

	q.SetShuffle(true)
	q.AddSongs(songs[0:2])
	q.songComplete()
	q.SetShuffle(false)
	wantSongs := append(songs[1:], songs[0:2]...)
	gotSongs := q.GetQueue()
	logDiff(t, wantSongs, gotSongs, "reversed shuffle")
}

func logDiff(t *testing.T, x, y interface{}, msg string) {

	diff := cmp.Diff(x, y)
	if diff != "" {
		t.Error(msg, diff)
	}
}

// Some benchmarks for shuffling/reversing queue.

// order of 30 µs / op
func BenchmarkQueue_SetShuffle_100(b *testing.B) {
	n := 100
	songs := testSongs()
	q := newQueue()

	for i := 0; i < n/len(songs); i++ {
		q.AddSongs(songs)
	}

	//b.Logf("queue size: %d", len(q.GetQueue()))

	for i := 0; i < b.N; i++ {
		q.SetShuffle(false)
		q.SetShuffle(true)
	}
}

// order of 350 µs / op
func BenchmarkQueue_SetShuffle_1000(b *testing.B) {
	n := 1000
	songs := testSongs()
	q := newQueue()

	for i := 0; i < n/len(songs); i++ {
		q.AddSongs(songs)
	}

	//b.Logf("queue size: %d", len(q.GetQueue()))

	for i := 0; i < b.N; i++ {
		q.SetShuffle(false)
		q.SetShuffle(true)
	}
}

// order of 5 ms / op
func BenchmarkQueue_SetShuffle_10000(b *testing.B) {
	n := 10000
	songs := testSongs()
	q := newQueue()

	for i := 0; i < n/len(songs); i++ {
		q.AddSongs(songs)
	}

	//b.Logf("queue size: %d", len(q.GetQueue()))

	for i := 0; i < b.N; i++ {
		q.SetShuffle(false)
		q.SetShuffle(true)
	}
}
