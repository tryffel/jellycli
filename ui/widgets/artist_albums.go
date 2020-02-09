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
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

const (
	albumCoverHeight   = 5
	albumCoverWidth    = 20
	albumCoverPaddingH = 1
	albumCoverPaddingW = 1
)

type ArtistHeader struct {
	*tview.Grid

	artist *models.Artist
	name   *tview.TextView

	prevBtn    *tview.Button
	infoBtn    *tview.Button
	anotherBtn *tview.Button
	similarBtn *tview.Button
	prevFunc   func()
}

func NewArtistHeader(prevFunc func()) *ArtistHeader {
	a := &ArtistHeader{
		Grid:       tview.NewGrid(),
		artist:     &models.Artist{},
		name:       tview.NewTextView(),
		prevBtn:    tview.NewButton("Back"),
		prevFunc:   prevFunc,
		anotherBtn: tview.NewButton("Play all"),
		similarBtn: tview.NewButton("Similar"),
	}

	a.name.SetBorderPadding(0, 0, 1, 1)
	a.name.SetText(a.artist.Name)
	a.name.SetText(fmt.Sprintf("%s\nAlbums: %d, Total: %s",
		a.artist.Name, a.artist.AlbumCount, util.SecToStringApproximate(a.artist.TotalDuration)))
	a.prevBtn.SetSelectedFunc(a.prevFunc)

	btns := []*tview.Button{a.prevBtn, a.anotherBtn, a.similarBtn}
	for _, v := range btns {
		v.SetBackgroundColor(config.ColorNavBarBtn)
		v.SetLabelColor(config.ColorPrimary)
	}

	a.Grid.SetRows(1, 1, 1, 1, 1, 1)
	a.Grid.SetColumns(1, 6, 2, 10, -1, 10, -1, 10, -3)
	a.Grid.SetMinSize(1, 6)
	a.Grid.SetBackgroundColor(config.ColorBackground)
	a.name.SetBackgroundColor(config.ColorBackground)
	a.name.SetBackgroundColor(config.ColorBackground)

	a.Grid.AddItem(a.prevBtn, 1, 1, 1, 1, 1, 5, false)
	a.Grid.AddItem(a.name, 1, 3, 2, 5, 1, 10, false)
	a.Grid.AddItem(a.anotherBtn, 4, 3, 1, 1, 1, 10, true)
	a.Grid.AddItem(a.similarBtn, 4, 5, 1, 1, 1, 10, false)
	return a
}

func (a *ArtistHeader) SetArtist(artist *models.Artist) {
	a.artist = artist
	a.name.SetText(fmt.Sprintf("%s\nAlbums: %d, Total: %s",
		a.artist.Name, a.artist.AlbumCount, util.SecToStringApproximate(a.artist.TotalDuration)))
}

//AlbumCover is a simple cover for album, it shows
// album name, year and possible artists
type AlbumCover struct {
	*tview.TextView
	album   *models.Album
	index   int
	name    string
	year    int
	artists []string
}

func NewAlbumCover(index int, album *models.Album) *AlbumCover {
	a := &AlbumCover{
		TextView: tview.NewTextView(),
		album:    album,
		index:    index,
	}

	a.SetBorder(false)
	a.SetBackgroundColor(config.ColorBackground)
	a.SetBorderPadding(0, 0, 1, 1)
	a.SetTextColor(config.ColorPrimary)
	ar := printArtists(a.artists, 40)
	text := fmt.Sprintf("%d. %s\n%d", index, album.Name, album.Year)
	if ar != "" {
		text += "\n" + ar
	}

	a.TextView.SetText(text)
	return a
}

func (a *AlbumCover) SetRect(x, y, w, h int) {
	_, _, currentW, currentH := a.GetRect()
	// todo: compact name & artists if necessary
	if currentH != h {
	}
	if currentW != w {
	}
	a.TextView.SetRect(x, y, w, h)
}

func (a *AlbumCover) SetSelected(selected bool) {
	if selected {
		a.SetTextColor(config.ColorSelection)
		a.SetBackgroundColor(config.ColorSelectionBackground)
	} else {
		a.SetTextColor(config.ColorPrimary)
		a.SetBackgroundColor(config.ColorBackground)
	}
}

//print multiple artists
func printArtists(artists []string, maxWidth int) string {
	var out string
	need := 0
	for i, v := range artists {
		need += len(v)
		if i > 0 {
			need += 2
		}
	}

	if need > maxWidth {
		out = fmt.Sprintf("%d artists", len(artists))
		if len(out) > maxWidth {
			return ""
		} else {
			return out
		}
	}

	for i, v := range artists {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}

//ArtisView as a view that contains
type ArtistView struct {
	*tview.Grid
	list        *twidgets.ScrollList
	header      *ArtistHeader
	listFocused bool
	selectFunc  func(album *models.Album)
	albumCovers []*AlbumCover
}

func (a *ArtistView) AddAlbum(c *AlbumCover) {
	a.list.AddItem(c)
	a.albumCovers = append(a.albumCovers, c)
}

func (a *ArtistView) Clear() {
	a.list.Clear()
	a.header.SetArtist(nil)
	a.albumCovers = make([]*AlbumCover, 0)
}

func (a *ArtistView) SetArtist(artist *models.Artist) {
	a.header.SetArtist(artist)
}

func (a *ArtistView) SetAlbums(albums []*models.Album) {
	a.list.Clear()
	a.albumCovers = make([]*AlbumCover, len(albums))

	items := make([]twidgets.ListItem, len(albums))
	for i, v := range albums {
		cover := NewAlbumCover(i, v)
		items[i] = cover
		a.albumCovers[i] = cover
	}
	a.list.AddItems(items...)
}

//NewArtistView constructs new artist view
func NewArtistView(selectAlbum func(album *models.Album)) *ArtistView {
	a := &ArtistView{
		Grid:       tview.NewGrid(),
		header:     NewArtistHeader(nil),
		selectFunc: selectAlbum,
	}
	a.list = twidgets.NewScrollList(a.selectAlbum)
	a.list.ItemHeight = 3

	a.SetBorder(true)
	a.SetBorderColor(config.ColorBorder)
	a.SetBackgroundColor(config.ColorBackground)
	a.list.SetBackgroundColor(config.ColorBackground)
	a.SetBorder(true)
	a.SetBorderColor(config.ColorBorder)

	a.Grid.SetRows(5, -1)
	a.Grid.SetColumns(-1)

	a.Grid.AddItem(a.header, 0, 0, 1, 1, 6, 25, false)
	a.Grid.AddItem(a.list, 1, 0, 1, 1, 6, 25, false)

	a.listFocused = false
	return a
}

func (a *ArtistView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if a.listFocused {
			index := a.list.GetSelectedIndex()
			if index == 0 && key == tcell.KeyUp {
				a.listFocused = false
				a.header.Focus(func(p tview.Primitive) {})
				a.list.Blur()
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

func (a *ArtistView) selectAlbum(index int) {
	if a.selectFunc != nil {
		index := a.list.GetSelectedIndex()
		album := a.albumCovers[index]
		a.selectFunc(album.album)
	}
}
