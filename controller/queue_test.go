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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &queue{
				items: tt.songs,
			}
			if got := q.GetQueue(); !reflect.DeepEqual(got, tt.songs) {
				t.Errorf("GetQueue() = %v, want %v", got, tt.songs)
			}
		})
	}
}

func Test_queue_Reorder(t *testing.T) {
	songs := testSongs()

	type ordering struct {
		from int
		to   int
	}

	tests := []struct {
		name      string
		songs     []*models.Song
		want      []*models.Song
		orderings []ordering
	}{
		{
			name:  "first-to-last",
			songs: songs,
			want: []*models.Song{
				songs[1], songs[2], songs[3], songs[4], songs[5], songs[6],
				songs[7], songs[8], songs[0],
			},
			orderings: []ordering{
				{0, 8},
			},
		},
		{
			name:  "last-to-first",
			songs: songs,
			want: []*models.Song{
				songs[8], songs[0], songs[1], songs[2], songs[3], songs[4],
				songs[5], songs[6], songs[7],
			},
			orderings: []ordering{
				{8, 0},
			},
		},
		/*
			{
				name:  "first-to-n, n-to-first",
				songs: songs,
				want: []*models.Song{
					songs[5], songs[1], songs[2], songs[3], songs[4],
					songs[0], songs[6], songs[7], songs[8],
				},
				orderings: []ordering{
					{0, 5},
					{4, 0},
				},
			},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &queue{
				items: make([]*models.Song, len(songs)),
			}
			copy(q.items, songs)
			for _, v := range tt.orderings {
				q.Reorder(v.from, v.to)
			}
			if got := q.GetQueue(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reorder() = %v, want %v", got, tt.want)
			}
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
			q := &queue{
				items:   tt.songs,
				history: []*models.Song{},
			}
			for i := 0; i < tt.complete; i++ {
				q.songComplete()
			}

			history := q.GetHistory(10)
			if !reflect.DeepEqual(history, tt.want) {
				t.Errorf("TestQueue songComplete history: got %v, want: %v", history, tt.want)
			}

			if !reflect.DeepEqual(q.GetQueue(), tt.songs[tt.complete:]) {
				t.Errorf("TestQueue songComplete remove items: got %v, want: %v",
					q.GetQueue(), tt.songs[tt.complete])
			}
		})
	}
}

func Test_queue_AddSongs(t *testing.T) {
	songs := testSongs()
	tests := []struct {
		songs []*models.Song
		name  string
		add   []*models.Song
		want  []*models.Song
	}{
		{
			songs: songs,
			add:   []*models.Song{songs[1], songs[2], songs[3]},
			want:  append(songs, songs[1], songs[2], songs[3]),
		},
		{
			songs: nil,
			add:   []*models.Song{songs[1], songs[2], songs[3]},
			want:  []*models.Song{songs[1], songs[2], songs[3]},
		},
		{
			songs: songs,
			add:   nil,
			want:  songs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &queue{
				items: tt.songs,
			}
			q.AddSongs(tt.add)
			got := q.GetQueue()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddSongs, got: %v, want: %v", got, tt.want)
			}

		})
	}
}

func Test_queue_QueueDuration(t *testing.T) {
	songs := testSongs()

	tests := []struct {
		songs []*models.Song
		name  string
		want  int
	}{
		{
			songs: songs,
			want:  611,
		},
		{
			songs: nil,
			want:  0,
		},
		{
			songs: []*models.Song{{Duration: 1}},
			want:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &queue{
				items: tt.songs,
			}
			if got := q.QueueDuration(); got != tt.want {
				t.Errorf("QueueDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
