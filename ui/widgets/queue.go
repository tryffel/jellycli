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
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

// Queue shows a list of songs similar to album.
// A history view is queue with reverse order and some little tweaks.
type Queue struct {
	*twidgets.Banner
	*previous
	list        *twidgets.ScrollList
	songs       []*albumSong
	listFocused bool

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)

	controller interfaces.QueueController

	description *tview.TextView
	prevBtn     *button
	clearBtn    *button
	prevFunc    func()
	clearFunc   func()
}

//NewQueue initializes new album view
func NewQueue() *Queue {
	q := &Queue{
		Banner:   twidgets.NewBanner(),
		list:     twidgets.NewScrollList(nil),
		previous: &previous{},

		description: tview.NewTextView(),
		prevBtn:     newButton("Back"),
		clearBtn:    newButton("Clear"),
	}

	q.list.ItemHeight = 2
	q.list.Padding = 0
	q.list.SetInputCapture(q.listHandler)
	q.list.SetBorder(true)
	q.list.SetBorderColor(config.Color.Border)
	q.list.Grid.SetColumns(1, -1)

	q.clearBtn.SetSelectedFunc(q.clearQueue)

	q.SetBorder(true)
	q.SetBorderColor(config.Color.Border)
	q.list.SetBackgroundColor(config.Color.Background)
	q.Grid.SetBackgroundColor(config.Color.Background)
	q.listFocused = false

	q.Banner.Grid.SetRows(1, 1, 1, 1, -1)
	q.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	q.Banner.Grid.SetMinSize(1, 6)

	q.Banner.Grid.AddItem(q.prevBtn, 0, 0, 1, 1, 1, 5, false)
	q.Banner.Grid.AddItem(q.description, 0, 2, 2, 6, 1, 10, false)
	q.Banner.Grid.AddItem(q.clearBtn, 3, 2, 1, 1, 1, 10, true)
	q.Banner.Grid.AddItem(q.list, 4, 0, 1, 8, 4, 10, false)

	btns := []*button{q.prevBtn, q.clearBtn}
	selectables := []twidgets.Selectable{q.prevBtn, q.clearBtn, q.list}
	for _, btn := range btns {
		btn.SetLabelColor(config.Color.ButtonLabel)
		btn.SetLabelColorActivated(config.Color.ButtonLabelSelected)
		btn.SetBackgroundColor(config.Color.ButtonBackground)
		btn.SetBackgroundColorActivated(config.Color.ButtonBackgroundSelected)
	}

	q.prevBtn.SetSelectedFunc(q.goBack)
	q.Banner.Selectable = selectables
	q.description.SetBackgroundColor(config.Color.Background)
	q.description.SetTextColor(config.Color.Text)
	q.printDescription()
	return q
}

// AddSong adds song to queue. If index is 0, add to beginning, if -1, add to end
func (q *Queue) AddSong(song *models.Song, index int) {
	var s *albumSong
	if index == -1 {
		s := newAlbumSong(song, false, len(q.songs))
		q.songs = append(q.songs, s)
	} else if index >= 0 || index < len(q.songs)-2 {
	}
	q.list.AddItem(s)
}

// SetSongs clears current songs and sets new ones
func (q *Queue) SetSongs(songs []*models.Song) {
	q.Clear()
	q.songs = make([]*albumSong, len(songs))
	items := make([]twidgets.ListItem, len(songs))
	for i, v := range songs {
		s := newAlbumSong(v, false, i+1)
		q.songs[i] = s
		q.songs[i].updateTextFunc = q.updateSongText
		items[i] = s
	}
	// colorize first item in queue
	if len(q.songs) > 0 {
		q.songs[0].playing = true
	}
	q.list.AddItems(items...)
	q.printDescription()
}

// Clear removes all songs
func (q *Queue) Clear() {
	q.list.Clear()
	q.songs = []*albumSong{}
	q.printDescription()
}

func (q *Queue) printDescription() {
	text := "Queue"
	if len(q.songs) > 0 {
		duration := 0
		for _, v := range q.songs {
			duration += v.song.Duration
		}
		text += fmt.Sprintf(": %d items\n%s", len(q.songs), util.SecToStringApproximate(duration))
	}
	q.description.SetText(text)
}

func (q *Queue) listHandler(key *tcell.EventKey) *tcell.EventKey {
	switch key.Key() {
	case tcell.KeyEnter:
		return nil
	case tcell.KeyCtrlJ:
		if q.controller != nil {
			index := q.list.GetSelectedIndex()
			_ = q.controller.Reorder(index, false)
		}
	case tcell.KeyCtrlK:
		if q.controller != nil {
			index := q.list.GetSelectedIndex()
			_ = q.controller.Reorder(index, true)
		}
	case tcell.KeyDEL:
		if q.controller != nil {
			index := q.list.GetSelectedIndex()
			q.controller.RemoveSong(index)
		}
	}
	return key
}

func (q *Queue) updateSongText(song *albumSong) {
	var name string
	if song.playing {
		song.SetTextColor(config.Color.TextSongPlaying)
	}

	if song.showDiscNum {
		name = fmt.Sprintf("%d %d. %s", song.song.DiscNumber, song.song.Index, song.song.Name)
	} else {
		name = fmt.Sprintf("%d. %s", song.index, song.song.Name)
	}

	text := song.getAlignedDuration(name)
	if len(song.song.Artists) > 0 {
		text += "\n     " + song.song.Artists[0].Name

	}
	song.SetText(text)
}

// call clearing queue
func (q *Queue) clearQueue() {
	if q.clearFunc != nil {
		q.clearFunc()
	}
}
