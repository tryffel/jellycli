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

package widgets

import (
	"fmt"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

type History struct {
	*Queue
}

func NewHistory() *History {
	h := &History{NewQueue()}
	return h
}

func (h *History) printDescription() {
	text := "History"
	if len(h.songs) > 0 {
		duration := 0
		for _, v := range h.songs {
			duration += v.song.Duration
		}
		text += fmt.Sprintf(": %d items\n%s", len(h.songs), util.SecToStringApproximate(duration))
	}
	h.description.SetText(text)
}

// Clear removes all songs
func (h *History) Clear() {
	h.list.Clear()
	h.songs = []*albumSong{}
	h.printDescription()
}

// SetSongs clears current songs and sets new ones
func (h *History) SetSongs(songs []*models.Song) {
	h.Clear()
	h.songs = make([]*albumSong, len(songs))
	items := make([]twidgets.ListItem, len(songs))
	for i, v := range songs {
		s := newAlbumSong(v, false, i+1)
		h.songs[i] = s
		h.songs[i].updateTextFunc = h.updateSongText
		items[i] = s
	}
	h.list.AddItems(items...)
	h.printDescription()
}
