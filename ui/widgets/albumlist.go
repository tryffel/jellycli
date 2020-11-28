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
	"tryffel.net/go/twidgets"
)

//ArtisView as a view that contains
type AlbumList struct {
	*itemList
	context       contextOperator
	paging        *PageSelector
	options       *dropDown
	pagingEnabled bool
	page          interfaces.Paging
	selectFunc    func(album *models.Album)
	albumCovers   []*AlbumCover

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
		var artist = ""
		if len(v.AdditionalArtists) > 0 {
			artist = v.AdditionalArtists[0].Name
		}
		text := fmt.Sprintf("%d. %s\n     %s - %d", offset+i+1, v.Name, artist, v.Year)
		cover.setText(text)
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
	a.EnablePaging(false)
	a.EnableSimilar(false)
	return a
}

func newFavoriteAlbums(selectAlbum func(album *models.Album), context contextOperator) *AlbumList {
	a := NewAlbumList(selectAlbum, context, nil, nil)
	a.EnablePaging(true)
	a.EnableSimilar(false)
	return a
}
