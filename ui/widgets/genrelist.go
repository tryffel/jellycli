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
	"gitlab.com/tslocum/cview"
	"strings"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/twidgets"
)

type GenreList struct {
	*itemList
	paging         *PageSelector
	selectFunc     func(genre models.IdName)
	selectPageFunc func(page interfaces.Paging)
	genres         []*Genre

	pagingEnabled bool
	page          interfaces.Paging
}

func NewGenreList() *GenreList {
	g := &GenreList{

		pagingEnabled: false,
		page:          interfaces.Paging{},
	}
	g.itemList = newItemList(g.selectGenre)

	g.paging = NewPageSelector(g.selectPage)
	g.list.Grid.SetColumns(1, -1)
	g.list.Padding = 1
	g.list.ItemHeight = 2

	g.pagingEnabled = true
	selectables := []twidgets.Selectable{g.prevBtn, g.paging.Previous, g.paging.Next, g.list}
	g.Banner.Selectable = selectables
	g.description.SetBackgroundColor(config.Color.Background)
	g.description.SetTextColor(config.Color.Text)

	g.reduceEnabled = true
	g.setReducerVisible = g.showReduceInput

	g.Banner.Grid.SetRows(1, 1, 1, 1, -1, 3)
	g.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	g.Banner.Grid.SetMinSize(1, 6)

	g.Banner.Grid.AddItem(g.prevBtn, 0, 0, 1, 1, 1, 5, false)
	g.Banner.Grid.AddItem(g.description, 0, 2, 2, 6, 1, 10, false)
	g.Banner.Grid.AddItem(g.paging, 3, 4, 1, 3, 1, 10, true)
	g.Banner.Grid.AddItem(g.list, 4, 0, 2, 8, 4, 10, false)
	return g
}

func (g *GenreList) Clear() {
	g.list.Clear()
	g.genres = make([]*Genre, 0)
	g.resetReduce()
}

func (g *GenreList) selectPage(n int) {
	if g.selectPageFunc != nil {
		g.paging.SetPage(n)
		g.page.CurrentPage = n
		g.selectPageFunc(g.page)
	}
}

func (g *GenreList) selectGenre(index int) {
	if g.selectFunc != nil {
		if index < len(g.genres) {
			genre := g.genres[index].genre
			g.selectFunc(*genre)
		}
	}
	g.resetReduce()
}

func (g *GenreList) SetPage(paging interfaces.Paging) {
	g.paging.SetPage(paging.CurrentPage)
	g.paging.SetTotalPages(paging.TotalPages)
	g.page = paging
	g.resetReduce()
}

func (g *GenreList) setGenres(genres []*models.IdName) {
	g.Clear()
	items := make([]twidgets.ListItem, len(genres))

	offset := 0
	if g.pagingEnabled {
		offset = g.page.Offset()
	}

	itemTexts := make([]string, len(genres))
	for i, v := range genres {
		genre := newGenre(v)
		genre.SetText(fmt.Sprintf("%d. %s", i+offset+1, v.Name))
		g.genres = append(g.genres, genre)
		items[i] = genre

		itemTexts[i] = strings.ToLower(v.Name)
	}

	g.items = items
	g.itemsTexts = itemTexts
	g.list.AddItems(items...)
}

func (g *GenreList) showReduceInput(visible bool) {
	if visible {
		g.Grid.AddItem(g.reduceInput, 5, 0, 1, 10, 1, 20, false)
		g.Grid.RemoveItem(g.list)
		g.Grid.AddItem(g.list, 4, 0, 1, 10, 6, 20, false)
	} else {
		g.Grid.RemoveItem(g.reduceInput)
		g.Grid.RemoveItem(g.list)
		g.Grid.AddItem(g.list, 4, 0, 2, 10, 6, 20, false)
	}
}

type Genre struct {
	*cview.TextView
	genre *models.IdName
}

func newGenre(genre *models.IdName) *Genre {
	g := &Genre{
		TextView: cview.NewTextView(),
		genre:    genre,
	}
	g.SetBackgroundColor(config.Color.Background)
	g.SetTextColor(config.Color.Text)

	g.SetText(genre.Name)
	return g
}

func (g *Genre) SetSelected(s twidgets.Selection) {
	if s == twidgets.Selected {
		g.SetTextColor(config.Color.TextSelected)
		g.SetBackgroundColor(config.Color.BackgroundSelected)
	} else if s == twidgets.Deselected {
		g.SetTextColor(config.Color.Text)
		g.SetBackgroundColor(config.Color.Background)
	} else if s == twidgets.Blurred {
		g.SetBackgroundColor(config.Color.TextDisabled)
	}
}
