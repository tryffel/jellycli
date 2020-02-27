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
	"github.com/rivo/uniseg"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

type button struct {
	*tview.Button
}

func (b *button) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.Button.InputHandler()
}

func (b *button) Focus(delegate func(p tview.Primitive)) {
	b.Button.Focus(delegate)
}

func (b *button) GetFocusable() tview.Focusable {
	return b.Button.GetFocusable()
}

func (b *button) SetBlurFunc(blur func(key tcell.Key)) {
	b.Button.SetBlurFunc(blur)
}

func newButton(label string) *button {
	return &button{
		Button: tview.NewButton(label),
	}

}

type albumSong struct {
	*tview.TextView
	song *models.Song
}

func (a *albumSong) SetSelected(selected twidgets.Selection) {
	switch selected {
	case twidgets.Selected:
		a.SetBackgroundColor(config.ColorBtnBackgroundSelected)
	case twidgets.Blurred:
		a.SetBackgroundColor(config.ColorBtnBackground)
	case twidgets.Deselected:
		a.SetBackgroundColor(config.ColorBackground)
	}
}

func (a *albumSong) SetRect(x, y, w, h int) {
	_, _, ch, cw := a.GetRect()
	a.TextView.SetRect(x, y, w, h)
	if cw != w && a.song != nil {
		a.setText()
	}
	if ch != h {
	}
}

func (a *albumSong) setText() {
	if a.song == nil {
		return
	}
	_, _, w, _ := a.GetRect()
	duration := util.SecToString(a.song.Duration)
	dL := len(duration)
	name := fmt.Sprintf("%d. %s", a.song.Index, a.song.Name)
	nameL := uniseg.GraphemeClusterCount(name)

	// width - duration - name - padding
	spaces := w - dL - nameL - 2
	space := ""

	if spaces <= 0 {
		lines := tview.WordWrap(name, w-2)
		if len(lines) >= 1 {
			name = lines[0] + "… "
		}
	} else {
		for {
			if len(space) == spaces {
				break
			}
			if len(space) < spaces-10 {
				space += "          "
			} else if len(space) < spaces-5 {
				space += "     "
			} else if len(space) < spaces-3 {
				space += "   "
			} else {
				space += " "
			}
		}
	}

	text := name + space + duration
	a.SetText(text)
}

func newAlbumSong(s *models.Song) *albumSong {
	song := &albumSong{
		TextView: tview.NewTextView(),
		song:     s,
	}
	song.SetBackgroundColor(config.ColorBackground)
	song.SetTextColor(config.ColorPrimary)
	song.setText()
	song.SetBorderPadding(0, 0, 1, 1)

	return song
}

// AlbumView shows user a header (album name, info, buttons) and list of songs
type AlbumView struct {
	*twidgets.Banner
	list        *twidgets.ScrollList
	songs       []*albumSong
	artist      *models.Artist
	album       *models.Album
	listFocused bool

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)

	description *tview.TextView
	prevBtn     *button
	infobtn     *button
	playBtn     *button
	prevFunc    func()
}

//NewAlbumView initializes new album view
func NewAlbumview(playSong func(song *models.Song), playSongs func(songs []*models.Song)) *AlbumView {
	a := &AlbumView{
		Banner:        twidgets.NewBanner(),
		list:          twidgets.NewScrollList(nil),
		playSongFunc:  playSong,
		playSongsFunc: playSongs,

		description: tview.NewTextView(),
		prevBtn:     newButton("Back"),
		infobtn:     newButton("Info"),
		playBtn:     newButton("Play"),
	}

	a.list.ItemHeight = 2
	a.list.Padding = 1
	a.list.SetInputCapture(a.listHandler)
	a.list.SetBorder(true)

	a.SetBorder(true)
	a.SetBorderColor(config.ColorBorder)
	a.list.SetBackgroundColor(config.ColorBackground)
	a.Grid.SetBackgroundColor(config.ColorBackground)
	a.listFocused = false
	a.playBtn.SetSelectedFunc(a.playAlbum)

	a.Banner.Grid.SetRows(1, 1, 1, 1, -1)
	a.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Banner.Grid.SetMinSize(1, 6)

	a.Banner.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Banner.Grid.AddItem(a.description, 0, 2, 2, 5, 1, 10, false)
	a.Banner.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, false)
	a.Banner.Grid.AddItem(a.infobtn, 3, 4, 1, 1, 1, 10, false)
	a.Banner.Grid.AddItem(a.list, 4, 0, 1, 8, 4, 10, false)

	btns := []*button{a.prevBtn, a.playBtn, a.infobtn}
	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.infobtn, a.list}
	for _, btn := range btns {
		btn.SetLabelColor(config.ColorBtnLabel)
		btn.SetLabelColorActivated(config.ColorBtnLabelSelected)
		btn.SetBackgroundColor(config.ColorBtnBackground)
		btn.SetBackgroundColorActivated(config.ColorBtnBackgroundSelected)
		btn.SetInputCapture(a.btnHandler)
	}
	a.Banner.Selectable = selectables
	a.description.SetBackgroundColor(config.ColorBackground)
	a.description.SetTextColor(config.ColorPrimary)
	return a
}

func (a *AlbumView) SetAlbum(album *models.Album, songs []*models.Song) {
	a.list.Clear()
	a.songs = make([]*albumSong, len(songs))
	items := make([]twidgets.ListItem, len(songs))

	album.SongCount = len(a.songs)
	a.album = album
	a.description.SetText(fmt.Sprintf("%s\n%d tracks  %s  %d",
		album.Name,
		album.SongCount, util.SecToStringApproximate(album.Duration), album.Year))
	for i, v := range songs {
		a.songs[i] = newAlbumSong(v)
		items[i] = a.songs[i]
	}

	a.list.AddItems(items...)
}

func (a *AlbumView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if a.listFocused {
			index := a.list.GetSelectedIndex()
			if index == 0 && (key == tcell.KeyUp || key == tcell.KeyCtrlK) {
				a.listFocused = false
				a.prevBtn.Focus(func(p tview.Primitive) {})
				a.list.Blur()
			} else if key == tcell.KeyEnter {
				a.playSong(index)
			} else {
				a.list.InputHandler()(event, setFocus)
			}
		} else {
			if key == tcell.KeyDown || key == tcell.KeyCtrlJ {
				a.listFocused = true
				a.list.Focus(func(p tview.Primitive) {})
			} else {
			}
		}
	}
}

func (a *AlbumView) playSong(index int) {
	if a.playSongFunc != nil {
		song := a.songs[index].song
		a.playSongFunc(song)
	}
}

func (a *AlbumView) playAlbum() {
	if a.playSongsFunc != nil {
		songs := make([]*models.Song, len(a.songs))
		for i, v := range a.songs {
			songs[i] = v.song
		}
		a.playSongsFunc(songs)
	}
}

func (a *AlbumView) Blur() {
	a.Grid.Blur()
}

// map other keys to tab to enable navigation between buttons without using tab
func (a *AlbumView) btnHandler(key *tcell.EventKey) *tcell.EventKey {
	switch key.Key() {
	case tcell.KeyCtrlJ, tcell.KeyDown:
		return tcell.NewEventKey(tcell.KeyTAB, ' ', tcell.ModNone)
	case tcell.KeyCtrlK, tcell.KeyUp:
		return tcell.NewEventKey(tcell.KeyBacktab, ' ', tcell.ModNone)
	default:
		return key
	}
}

func (a *AlbumView) listHandler(key *tcell.EventKey) *tcell.EventKey {
	btn := a.btnHandler(key)
	if btn != key {
		return btn
	}
	if key.Key() == tcell.KeyEnter {
		index := a.list.GetSelectedIndex()
		a.playSong(index)
		return nil
	}
	return key
}
