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

type albumHeader struct {
	*tview.Grid
	artist      *models.Artist
	album       *models.Album
	description *tview.TextView

	prevBtn  *tview.Button
	infobtn  *tview.Button
	playBtn  *tview.Button
	prevFunc func()
}

func newAlbumHeader(prevFunc func()) *albumHeader {
	a := &albumHeader{
		Grid:        tview.NewGrid(),
		artist:      nil,
		album:       nil,
		description: tview.NewTextView(),
		prevBtn:     tview.NewButton("Back"),
		infobtn:     tview.NewButton("Info"),
		playBtn:     tview.NewButton("Play"),
		prevFunc:    prevFunc,
	}

	a.description.SetBorderPadding(0, 0, 1, 1)

	btns := []*tview.Button{a.prevBtn, a.infobtn, a.playBtn}
	for _, v := range btns {
		v.SetBackgroundColor(config.ColorBtnBackground)
		v.SetLabelColor(config.ColorBtnLabel)
		v.SetBackgroundColorActivated(config.ColorBtnBackgroundSelected)
		v.SetLabelColorActivated(config.ColorBtnLabelSelected)
	}

	a.SetBorder(true)
	a.SetBorderColor(config.ColorBorder)

	a.Grid.SetRows(1, 1, 1, 1)
	a.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Grid.SetMinSize(1, 6)
	a.Grid.SetBackgroundColor(config.ColorBackground)

	a.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Grid.AddItem(a.description, 0, 2, 2, 5, 1, 10, false)
	a.Grid.AddItem(a.infobtn, 3, 4, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, true)

	a.description.SetBackgroundColor(config.ColorBackground)
	a.description.SetTextColor(config.ColorPrimary)

	return a
}

func (a *albumHeader) SetArtist(artist *models.Artist) {
	a.artist = artist
}

func (a *albumHeader) SetAlbum(album *models.Album) {
	a.album = album
	a.description.SetText(fmt.Sprintf("%s\n%d tracks  %s  %d",
		album.Name,
		album.SongCount, util.SecToStringApproximate(album.Duration), album.Year))
}

type albumSong struct {
	*tview.TextView
	song *models.Song
}

func (a *albumSong) SetSelected(selected bool) {
	if selected {
		a.SetBackgroundColor(config.ColorProgress)
	} else {
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

type AlbumView struct {
	*tview.Grid
	list        *twidgets.ScrollList
	songs       []*albumSong
	header      *albumHeader
	listFocused bool

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)
}

func NewAlbumview(playSong func(song *models.Song), playSongs func(songs []*models.Song)) *AlbumView {
	a := &AlbumView{
		Grid:          tview.NewGrid(),
		list:          twidgets.NewScrollList(nil),
		header:        newAlbumHeader(nil),
		playSongFunc:  playSong,
		playSongsFunc: playSongs,
	}

	a.list.ItemHeight = 2
	a.list.Padding = 1

	a.SetBorder(true)
	a.SetBorderColor(config.ColorBorder)
	a.list.SetBackgroundColor(config.ColorBackground)
	a.Grid.SetBackgroundColor(config.ColorBackground)
	a.Grid.SetRows(6, -1)
	a.Grid.SetColumns(-1)

	a.Grid.AddItem(a.header, 0, 0, 1, 1, 5, 25, false)
	a.Grid.AddItem(a.list, 1, 0, 1, 1, 5, 25, false)
	a.listFocused = false

	a.header.playBtn.SetSelectedFunc(a.playAlbum)
	return a
}

func (a *AlbumView) SetAlbum(album *models.Album, songs []*models.Song) {
	a.list.Clear()
	a.songs = make([]*albumSong, len(songs))
	items := make([]twidgets.ListItem, len(songs))

	album.SongCount = len(a.songs)
	a.header.SetAlbum(album)
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
			if index == 0 && key == tcell.KeyUp {
				a.listFocused = false
				a.header.Focus(func(p tview.Primitive) {})
				a.list.Blur()
			} else if key == tcell.KeyEnter {
				a.playSong(index)
			} else {
				a.list.InputHandler()(event, setFocus)
			}
		} else {
			r := event.Rune()
			if r == 'j' || key == tcell.KeyDown {
				a.listFocused = true
				a.header.Blur()
				a.list.Focus(func(p tview.Primitive) {})
			} else {
				a.header.InputHandler()(event, setFocus)
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
