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
	h.printDescription()
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
