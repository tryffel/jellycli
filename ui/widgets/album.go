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
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"
	"tryffel.net/go/twidgets"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/models"
)

type albumHeader struct {
	*tview.Grid
	artist      *models.Artist
	album       *models.Album
	name        *tview.TextView
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
		name:        tview.NewTextView(),
		description: tview.NewTextView(),
		prevBtn:     tview.NewButton("Back"),
		infobtn:     tview.NewButton("Info"),
		playBtn:     tview.NewButton("Play"),
		prevFunc:    prevFunc,
	}

	a.name.SetBorderPadding(0, 0, 1, 1)
	a.description.SetBorderPadding(0, 0, 1, 1)

	btns := []*tview.Button{a.prevBtn, a.infobtn, a.playBtn}
	for _, v := range btns {
		v.SetBackgroundColor(config.ColorNavBarBtn)
		v.SetLabelColor(config.ColorPrimary)
	}

	a.Grid.SetRows(1, 1, 1, 1, 1, 1)
	a.Grid.SetColumns(1, 6, 2, 10, -1, 10, -1, 10, -3)
	a.Grid.SetMinSize(1, 6)
	a.Grid.SetBackgroundColor(config.ColorBackground)

	a.Grid.AddItem(a.prevBtn, 1, 1, 1, 1, 1, 5, true)
	a.Grid.AddItem(a.name, 1, 3, 1, 5, 1, 10, false)
	a.Grid.AddItem(a.description, 2, 3, 1, 5, 1, 10, false)
	a.Grid.AddItem(a.infobtn, 4, 5, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.playBtn, 4, 3, 1, 1, 1, 10, false)

	return a
}

func (a *albumHeader) SetArtist(artist *models.Artist) {
	a.artist = artist
}

func (a *albumHeader) SetAlbum(album *models.Album) {
	a.album = album
	a.name.SetText(album.Name)
	a.description.SetText(fmt.Sprintf("%d tracks  %s  %d",
		album.SongCount, SecToStringApproximate(album.Duration), album.Year))
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
	duration := SecToString(a.song.Duration)
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
	song.setText()
	song.SetBorderPadding(0, 0, 1, 1)

	return song
}

type AlbumView struct {
	*tview.Grid
	list   *twidgets.ScrollList
	songs  []*albumSong
	header *albumHeader
}

func NewAlbumview() *AlbumView {
	a := &AlbumView{
		Grid:   tview.NewGrid(),
		list:   twidgets.NewScrollList(nil),
		header: newAlbumHeader(nil),
	}

	a.list.ItemHeight = 2
	a.list.Padding = 0

	a.Grid.SetRows(5, -1)
	a.Grid.SetColumns(-1)

	a.Grid.AddItem(a.header, 0, 0, 1, 1, 6, 25, false)
	a.Grid.AddItem(a.list, 1, 0, 1, 1, 6, 25, true)

	return a
}

func (a *AlbumView) SetAlbum(album *models.Album, songs []*models.Song) {
	a.list.Clear()
	a.songs = make([]*albumSong, len(songs))
	a.header.SetAlbum(album)
	for i, v := range songs {
		a.songs[i] = newAlbumSong(v)
		a.list.AddItem(a.songs[i])
	}
}
