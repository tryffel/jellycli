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
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

//AlbumCover is a simple cover for album, it shows
// album name, year and possible artists
type AlbumCover struct {
	*cview.TextView
	album   *models.Album
	index   int
	name    string
	year    int
	artists []string
}

func NewAlbumCover(index int, album *models.Album) *AlbumCover {
	a := &AlbumCover{
		TextView: cview.NewTextView(),
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
	*itemList
	context       contextOperator
	paging        *PageSelector
	options       *dropDown
	pagingEnabled bool
	page          interfaces.Paging
	artistMode    bool
	selectFunc    func(album *models.Album)
	albumCovers   []*AlbumCover

	artist *models.Artist

	infoBtn        *button
	playBtn        *button
	selectPageFunc func(paging interfaces.Paging)
	similarFunc    func(id models.Id)
	similarEnabled bool

	sort          *sort
	filter        *filter
	filterBtn     *button
	filterEnabled bool
	queryOpts     *interfaces.QueryOpts
	queryFunc     func(opts *interfaces.QueryOpts)
}

func (a *AlbumList) AddAlbum(c *AlbumCover) {
	a.list.AddItem(c)
	a.albumCovers = append(a.albumCovers, c)
}

func (a *AlbumList) Clear() {
	a.list.Clear()
	//a.SetArtist(nil)
	a.albumCovers = make([]*AlbumCover, 0)

	if a.filterBtn != nil {
		a.filter.Clear()
	}
}

// SetPlaylists sets albumList cover
func (a *AlbumList) SetArtist(artist *models.Artist) {
	a.artist = artist
	if artist != nil {
		favorite := ""
		if artist.Favorite {
			favorite = charFavorite + " "
		}

		a.description.SetText(fmt.Sprintf("%s%s\nAlbums: %d, Total: %s",
			favorite, a.artist.Name, a.artist.AlbumCount, util.SecToStringApproximate(a.artist.TotalDuration)))
	} else {
		a.description.SetText("")
	}
}

func (a *AlbumList) SetLabel(label string) {
	a.description.SetText(label)
}

func (a *AlbumList) SetText(text string) {
	a.description.SetText(text)
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
	a.queryOpts.Paging = a.page
	if a.queryFunc != nil {
		a.queryFunc(a.queryOpts)
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
	a.setButtons()
}

func (a *AlbumList) EnableSimilar(enabled bool) {
	a.similarEnabled = enabled
	a.setButtons()
}

func (a *AlbumList) EnableFilter(enabled bool) {
	if a.filterBtn != nil {
		a.filterEnabled = enabled
		if enabled {
			a.similarEnabled = false
		}
		a.setButtons()
	}
}

func (a *AlbumList) setButtons() {
	a.Banner.Grid.Clear()
	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn}
	a.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Grid.AddItem(a.description, 0, 2, 2, 6, 1, 10, false)
	a.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, false)

	if a.pagingEnabled {
		selectables = append(selectables, a.paging.Previous, a.paging.Next)
		a.Grid.AddItem(a.paging, 3, 4, 1, 3, 1, 10, false)
	}
	if a.similarEnabled {
		selectables = append(selectables, a.options)
		col := 4
		if a.pagingEnabled {
			col = 6
		}
		a.Grid.AddItem(a.options, 3, col, 1, 1, 1, 10, false)
	} else if a.filterEnabled && a.filterBtn != nil {
		selectables = append(selectables, a.sort, a.filterBtn)
		col := 4
		if a.pagingEnabled {
			col = 6
		}
		a.Grid.AddItem(a.sort, 3, col, 1, 1, 1, 10, false)
		a.Grid.AddItem(a.filterBtn, 3, col+2, 1, 1, 1, 10, false)
	}

	selectables = append(selectables, a.list)
	a.Banner.Selectable = selectables
	a.Grid.AddItem(a.list, 4, 0, 1, 10, 6, 20, false)
}

//NewAlbumList constructs new albumList view
func NewAlbumList(selectAlbum func(album *models.Album), context contextOperator,
	queryFunc func(opts *interfaces.QueryOpts), filterFunc openFilterFunc) *AlbumList {
	a := &AlbumList{
		itemList:   newItemList(nil),
		context:    context,
		selectFunc: selectAlbum,
		artist:     &models.Artist{},
		playBtn:    newButton("Play all"),
		options:    newDropDown("Options"),

		queryFunc: queryFunc,
		queryOpts: interfaces.DefaultQueryOpts(),
	}
	a.paging = NewPageSelector(a.selectPage)
	a.list = newScrollList(a.selectAlbum)
	a.list.ItemHeight = 3

	a.list.Grid.SetColumns(-1, 5)

	if queryFunc != nil {
		a.sort = newSort(a.setSorting,
			interfaces.SortByName,
			interfaces.SortByArtist,
			interfaces.SortByDate,
			interfaces.SortByRandom,
			interfaces.SortByPlayCount,
		)
	}

	a.filter = newFilter("album", a.setFilter, a.filterApplied)
	if filterFunc != nil {
		a.filterBtn = newButton("Filter")
		a.filterBtn.SetSelectedFunc(func() {
			filterFunc(a.filter, nil)
		})
	}
	a.filterEnabled = true

	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.options,
		a.paging.Previous, a.paging.Next, a.list}
	a.Banner.Selectable = selectables

	a.Grid.SetRows(1, 1, 1, 1, -1)
	a.Grid.SetColumns(6, 2, 10, -1, 10, -1, 15, -1, 10, -3)
	a.Grid.SetMinSize(1, 6)
	a.Grid.SetBackgroundColor(config.Color.Background)
	a.list.Grid.SetColumns(1, -1)

	a.listFocused = false
	a.pagingEnabled = true
	a.similarEnabled = true

	if a.context != nil {
		a.list.AddContextItem("Instant mix", 0, func(index int) {
			if index < len(a.albumCovers) && a.context != nil {
				album := a.albumCovers[index]
				a.context.InstantMix(album.album)
			}
		})
		a.options.AddOption("Instant Mix", func() {
			a.context.InstantMix(a.artist)
		})
		a.options.AddOption("Show similar", func() {
			if a.similarEnabled {
				a.showSimilar()
			}
		})
		a.options.AddOption("Show in browser", func() {
			a.context.OpenInBrowser(a.artist)
		})
	}

	a.setButtons()
	return a
}

func (a *AlbumList) filterApplied(status bool) {
	if status {
		a.filterBtn.SetLabel("Filter *")
	} else {
		a.filterBtn.SetLabel("Filter")
	}
}

func (a *AlbumList) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			a.list.InputHandler()(event, setFocus)
		} else {
			a.Banner.InputHandler()(event, setFocus)
		}
	}
}

func (a *AlbumList) selectAlbum(index int) {
	if a.selectFunc != nil {
		index := a.list.GetSelectedIndex()
		if len(a.albumCovers) > index {
			album := a.albumCovers[index]
			a.selectFunc(album.album)
		}
	}
}

func (a *AlbumList) showSimilar() {
	if a.similarFunc != nil {
		a.similarFunc(a.artist.Id)
	}
}

func (a *AlbumList) setSorting(sort interfaces.Sort) {
	a.queryOpts.Sort = sort
	if a.queryFunc != nil {
		a.queryFunc(a.queryOpts)
	}
}

func (a *AlbumList) setFilter(filter interfaces.Filter) {
	a.queryOpts.Filter = filter
	if a.queryFunc != nil {
		a.queryFunc(a.queryOpts)
	}
}

func newLatestAlbums(selectAlbum func(album *models.Album), context contextOperator) *AlbumList {
	a := NewAlbumList(selectAlbum, context, nil, nil)
	a.EnableArtistMode(false)
	a.EnablePaging(false)
	a.EnableSimilar(false)
	return a
}

func newFavoriteAlbums(selectAlbum func(album *models.Album), context contextOperator) *AlbumList {
	a := NewAlbumList(selectAlbum, context, nil, nil)
	a.EnableArtistMode(false)
	a.EnablePaging(true)
	a.EnableSimilar(false)
	return a
}
