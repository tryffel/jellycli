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
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/twidgets"
)

type GenreList struct {
	*twidgets.Banner
	list *twidgets.ScrollList
	*previous
	paging         *PageSelector
	selectFunc     func(genre models.IdName)
	selectPageFunc func(page interfaces.Paging)
	genres         []*Genre

	listFocused   bool
	description   *cview.TextView
	backBtn       *button
	pagingEnabled bool
	page          interfaces.Paging
}

func NewGenreList() *GenreList {
	g := &GenreList{
		Banner:        twidgets.NewBanner(),
		previous:      &previous{},
		listFocused:   false,
		description:   cview.NewTextView(),
		backBtn:       newButton("Back"),
		pagingEnabled: false,
		page:          interfaces.Paging{},
	}

	g.paging = NewPageSelector(g.selectPage)
	g.list = twidgets.NewScrollList(g.selectGenre)

	g.list.SetBorder(true)
	g.list.SetBorderColor(config.Color.Border)
	g.list.Grid.SetColumns(1, -1)

	g.list.SetBackgroundColor(config.Color.Background)
	g.Grid.SetBackgroundColor(config.Color.Background)
	g.list.Padding = 1
	g.list.ItemHeight = 2

	g.SetBorder(true)
	g.SetBorderColor(config.Color.Border)
	g.SetBackgroundColor(config.Color.Background)

	g.pagingEnabled = true
	btns := []*button{g.backBtn, g.paging.Previous, g.paging.Next}
	selectables := []twidgets.Selectable{g.backBtn, g.paging.Previous, g.paging.Next, g.list}
	for _, btn := range btns {
		btn.SetLabelColor(config.Color.ButtonLabel)
		btn.SetLabelColorFocused(config.Color.ButtonLabelSelected)
		btn.SetBackgroundColor(config.Color.ButtonBackground)
		btn.SetBackgroundColorFocused(config.Color.ButtonBackgroundSelected)
	}

	g.backBtn.SetSelectedFunc(g.goBack)
	g.Banner.Selectable = selectables
	g.description.SetBackgroundColor(config.Color.Background)
	g.description.SetTextColor(config.Color.Text)

	g.Banner.Grid.SetRows(1, 1, 1, 1, -1)
	g.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	g.Banner.Grid.SetMinSize(1, 6)

	g.Banner.Grid.AddItem(g.backBtn, 0, 0, 1, 1, 1, 5, false)
	g.Banner.Grid.AddItem(g.description, 0, 2, 2, 6, 1, 10, false)
	g.Banner.Grid.AddItem(g.paging, 3, 4, 1, 3, 1, 10, true)
	g.Banner.Grid.AddItem(g.list, 4, 0, 1, 8, 4, 10, false)
	return g
}

func (g *GenreList) Clear() {
	g.list.Clear()
	g.genres = make([]*Genre, 0)
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
}

func (g *GenreList) SetPage(paging interfaces.Paging) {
	g.paging.SetPage(paging.CurrentPage)
	g.paging.SetTotalPages(paging.TotalPages)
	g.page = paging
}

func (g *GenreList) setGenres(genres []*models.IdName) {
	g.Clear()
	items := make([]twidgets.ListItem, len(genres))

	offset := 0
	if g.pagingEnabled {
		offset = g.page.Offset()
	}
	for i, v := range genres {
		genre := newGenre(v)
		genre.SetText(fmt.Sprintf("%d. %s", i+offset+1, v.Name))
		g.genres = append(g.genres, genre)
		items[i] = genre
	}

	g.list.AddItems(items...)
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
