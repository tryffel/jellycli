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

func (a *AlbumCover) setText(text string) {
	a.TextView.SetText(text)
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
type AlbumList struct {
	*twidgets.Banner
	*previous
	paging        *PageSelector
	pagingEnabled bool
	page          interfaces.Paging
	artistMode    bool
	list          *twidgets.ScrollList
	listFocused   bool
	selectFunc    func(album *models.Album)
	albumCovers   []*AlbumCover

	artist *models.Artist
	name   *tview.TextView

	prevBtn        *button
	infoBtn        *button
	playBtn        *button
	similarBtn     *button
	prevFunc       func()
	selectPageFunc func(paging interfaces.Paging)
}

func (a *AlbumList) AddAlbum(c *AlbumCover) {
	a.list.AddItem(c)
	a.albumCovers = append(a.albumCovers, c)
}

func (a *AlbumList) Clear() {
	a.list.Clear()
	//a.SetArtist(nil)
	a.albumCovers = make([]*AlbumCover, 0)
}

// SetPlaylists sets albumList cover
func (a *AlbumList) SetArtist(artist *models.Artist) {
	a.artist = artist
	if artist != nil {
		a.name.SetText(fmt.Sprintf("%s\nAlbums: %d, Total: %s",
			a.artist.Name, a.artist.AlbumCount, util.SecToStringApproximate(a.artist.TotalDuration)))
	} else {
		a.name.SetText("")
	}
}

func (a *AlbumList) SetText(text string) {
	a.name.SetText(text)
}

func (a *AlbumList) SetPage(paging interfaces.Paging) {
	a.paging.SetPage(paging.CurrentPage)
	a.paging.SetTotalPages(paging.TotalPages)
	a.page = paging
}

// EnableArtistMode enabled single albumList mode. If disabled, albums don't necessarily have same albumList
// and are formatted differently. This does not update content.
func (a *AlbumList) EnableArtistMode(enabled bool) {
	a.artistMode = enabled
}

func (a *AlbumList) selectPage(n int) {
	a.paging.SetPage(n)
	a.page.CurrentPage = n
	if a.selectPageFunc != nil {
		a.selectPageFunc(a.page)
	}
}

// SetPlaylist sets albums
func (a *AlbumList) SetAlbums(albums []*models.Album) {
	a.list.Clear()
	a.albumCovers = make([]*AlbumCover, len(albums))

	offset := 0
	if a.pagingEnabled {
		offset = a.page.Offset()
	}

	items := make([]twidgets.ListItem, len(albums))
	for i, v := range albums {
		cover := NewAlbumCover(offset+i+1, v)
		items[i] = cover
		a.albumCovers[i] = cover
		if !a.artistMode {
			var artist = ""
			if len(v.AdditionalArtists) > 0 {
				artist = v.AdditionalArtists[0].Name
			}
			text := fmt.Sprintf("%d. %s\n     %s - %d", offset+i+1, v.Name, artist, v.Year)
			cover.setText(text)
		}
	}
	a.list.AddItems(items...)
}

// EnablePaging enables paging and shows page on banner
func (a *AlbumList) EnablePaging(enabled bool) {
	if a.pagingEnabled && enabled {
		return
	}
	if !a.pagingEnabled && !enabled {
		return
	}
	a.pagingEnabled = enabled
	if enabled {
		selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.paging.Previous, a.paging.Next, a.list}
		a.Banner.Selectable = selectables
		a.Banner.Grid.AddItem(a.paging, 3, 4, 1, 3, 1, 10, true)
	} else {
		selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.list}
		a.Banner.Selectable = selectables
		a.Banner.Grid.RemoveItem(a.paging)
		a.page.CurrentPage = 0
	}
}

//NewAlbumList constructs new albumList view
func NewAlbumList(selectAlbum func(album *models.Album)) *AlbumList {
	a := &AlbumList{
		Banner:     twidgets.NewBanner(),
		previous:   &previous{},
		selectFunc: selectAlbum,
		artist:     &models.Artist{},
		name:       tview.NewTextView(),
		prevBtn:    newButton("Back"),
		prevFunc:   nil,
		playBtn:    newButton("Play all"),
		similarBtn: newButton("Similar"),
	}
	a.paging = NewPageSelector(a.selectPage)
	a.list = twidgets.NewScrollList(a.selectAlbum)
	a.list.ItemHeight = 3

	a.SetBorder(true)
	a.SetBorderColor(config.Color.Border)
	a.SetBackgroundColor(config.Color.Background)
	a.list.SetBackgroundColor(config.Color.Background)
	a.list.SetBorder(true)
	a.list.SetBorderColor(config.Color.Border)
	a.list.Grid.SetColumns(-1, 5)
	a.SetBorderColor(config.Color.Border)

	btns := []*button{a.prevBtn, a.playBtn, a.similarBtn, a.paging.Previous, a.paging.Next}
	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.paging.Previous, a.paging.Next, a.list}
	for _, v := range btns {
		v.SetBackgroundColor(config.Color.ButtonBackground)
		v.SetLabelColor(config.Color.ButtonLabel)
		v.SetBackgroundColorActivated(config.Color.ButtonBackgroundSelected)
		v.SetLabelColorActivated(config.Color.ButtonLabelSelected)
	}

	a.prevBtn.SetSelectedFunc(a.goBack)

	a.Banner.Selectable = selectables

	a.Grid.SetRows(1, 1, 1, 1, -1)
	a.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Grid.SetMinSize(1, 6)
	a.Grid.SetBackgroundColor(config.Color.Background)
	a.name.SetBackgroundColor(config.Color.Background)
	a.name.SetTextColor(config.Color.Text)

	a.list.Grid.SetColumns(1, -1)

	a.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Grid.AddItem(a.name, 0, 2, 2, 6, 1, 10, false)
	a.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.paging, 3, 4, 1, 3, 1, 10, false)
	//a.Grid.AddItem(a.similarBtn, 3, 4, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.list, 4, 0, 1, 8, 6, 20, false)

	a.listFocused = false
	a.pagingEnabled = true
	return a
}

func (a *AlbumList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		a.Banner.InputHandler()(event, setFocus)
	}
}

func (a *AlbumList) selectAlbum(index int) {
	if a.selectFunc != nil {
		index := a.list.GetSelectedIndex()
		album := a.albumCovers[index]
		a.selectFunc(album.album)
	}
}
