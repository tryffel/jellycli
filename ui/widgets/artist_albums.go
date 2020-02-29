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
	a.SetBackgroundColor(config.Color.Background)
	a.SetBorderPadding(0, 0, 1, 1)
	a.SetTextColor(config.Color.Text)
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

func (a *AlbumCover) SetSelected(selected twidgets.Selection) {
	switch selected {
	case twidgets.Selected:
		a.SetBackgroundColor(config.Color.BackgroundSelected)
		a.SetTextColor(config.Color.TextSelected)
	case twidgets.Blurred:
		a.SetBackgroundColor(config.Color.TextDisabled)
	case twidgets.Deselected:
		a.SetBackgroundColor(config.Color.Background)
		a.SetTextColor(config.Color.Text)
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
	*twidgets.Banner
	list        *twidgets.ScrollList
	listFocused bool
	selectFunc  func(album *models.Album)
	albumCovers []*AlbumCover

	artist *models.Artist
	name   *tview.TextView

	prevBtn    *button
	infoBtn    *button
	playBtn    *button
	similarBtn *button
	prevFunc   func()
}

func (a *ArtistView) AddAlbum(c *AlbumCover) {
	a.list.AddItem(c)
	a.albumCovers = append(a.albumCovers, c)
}

func (a *ArtistView) Clear() {
	a.list.Clear()
	a.SetArtist(nil)
	a.artist = nil
	a.albumCovers = make([]*AlbumCover, 0)
}

// SetArtist sets artist cover
func (a *ArtistView) SetArtist(artist *models.Artist) {
	a.artist = artist
	if artist != nil {
		a.name.SetText(fmt.Sprintf("%s\nAlbums: %d, Total: %s",
			a.artist.Name, a.artist.AlbumCount, util.SecToStringApproximate(a.artist.TotalDuration)))
	} else {
		a.name.SetText("")
	}
}

// SetAlbum sets albums
func (a *ArtistView) SetAlbums(albums []*models.Album) {
	a.list.Clear()
	a.albumCovers = make([]*AlbumCover, len(albums))

	items := make([]twidgets.ListItem, len(albums))
	for i, v := range albums {
		cover := NewAlbumCover(i+1, v)
		items[i] = cover
		a.albumCovers[i] = cover
	}
	a.list.AddItems(items...)
}

//NewArtistView constructs new artist view
func NewArtistView(selectAlbum func(album *models.Album)) *ArtistView {
	a := &ArtistView{
		Banner:     twidgets.NewBanner(),
		selectFunc: selectAlbum,
		artist:     &models.Artist{},
		name:       tview.NewTextView(),
		prevBtn:    newButton("Back"),
		prevFunc:   nil,
		playBtn:    newButton("Play all"),
		similarBtn: newButton("Similar"),
	}
	a.list = twidgets.NewScrollList(a.selectAlbum)
	a.list.ItemHeight = 3

	a.SetBorder(true)
	a.SetBorderColor(config.Color.Border)
	a.SetBackgroundColor(config.Color.Background)
	a.list.SetBackgroundColor(config.Color.Background)
	a.list.SetBorder(true)
	a.list.SetBorderColor(config.Color.Border)
	a.SetBorderColor(config.Color.Border)

	btns := []*button{a.prevBtn, a.playBtn, a.similarBtn}
	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.similarBtn, a.list}
	for _, v := range btns {
		v.SetBackgroundColor(config.Color.ButtonBackground)
		v.SetLabelColor(config.Color.ButtonLabel)
		v.SetBackgroundColorActivated(config.Color.ButtonBackgroundSelected)
		v.SetLabelColorActivated(config.Color.ButtonLabelSelected)
	}

	a.Banner.Selectable = selectables

	a.Grid.SetRows(1, 1, 1, 1, -1)
	a.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Grid.SetMinSize(1, 6)
	a.Grid.SetBackgroundColor(config.Color.Background)
	a.name.SetBackgroundColor(config.Color.Background)
	a.name.SetTextColor(config.Color.Text)

	a.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Grid.AddItem(a.name, 0, 2, 2, 5, 1, 10, false)
	a.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.similarBtn, 3, 4, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.list, 4, 0, 1, 8, 6, 20, false)

	a.listFocused = false
	return a
}

func (a *ArtistView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		a.Banner.InputHandler()(event, setFocus)
	}
}

func (a *ArtistView) selectAlbum(index int) {
	if a.selectFunc != nil {
		index := a.list.GetSelectedIndex()
		album := a.albumCovers[index]
		a.selectFunc(album.album)
	}
}
