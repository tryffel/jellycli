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
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

type ArtistList struct {
	*twidgets.Banner
	list *twidgets.ScrollList
	*previous
	paging         *PageSelector
	selectFunc     func(artist *models.Artist)
	selectPageFunc func(page interfaces.Paging)
	artists        []*ArtistCover

	listFocused   bool
	description   *cview.TextView
	backBtn       *button
	pagingEnabled bool
	page          interfaces.Paging
}

func NewArtistList(selectFunc func(artist *models.Artist)) *ArtistList {
	a := &ArtistList{
		Banner:      twidgets.NewBanner(),
		selectFunc:  selectFunc,
		artists:     make([]*ArtistCover, 0),
		previous:    &previous{},
		description: cview.NewTextView(),
		backBtn:     newButton("Back"),
	}
	a.paging = NewPageSelector(a.selectPage)
	a.list = twidgets.NewScrollList(a.selectArtist)

	a.list.SetInputCapture(a.listHandler)
	a.list.SetBorder(true)
	a.list.SetBorderColor(config.Color.Border)
	a.list.Grid.SetColumns(1, -1)

	a.list.SetBackgroundColor(config.Color.Background)
	a.Grid.SetBackgroundColor(config.Color.Background)
	a.list.Padding = 1
	a.list.ItemHeight = 2

	a.SetBorder(true)
	a.SetBorderColor(config.Color.Border)
	a.SetBackgroundColor(config.Color.Background)
	a.SetBorder(true)

	a.pagingEnabled = true
	btns := []*button{a.backBtn, a.paging.Previous, a.paging.Next}
	selectables := []twidgets.Selectable{a.backBtn, a.paging.Previous, a.paging.Next, a.list}
	for _, btn := range btns {
		btn.SetLabelColor(config.Color.ButtonLabel)
		btn.SetLabelColorActivated(config.Color.ButtonLabelSelected)
		btn.SetBackgroundColor(config.Color.ButtonBackground)
		btn.SetBackgroundColorActivated(config.Color.ButtonBackgroundSelected)
	}

	a.backBtn.SetSelectedFunc(a.goBack)
	a.Banner.Selectable = selectables
	a.description.SetBackgroundColor(config.Color.Background)
	a.description.SetTextColor(config.Color.Text)

	a.Banner.Grid.SetRows(1, 1, 1, 1, -1)
	a.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Banner.Grid.SetMinSize(1, 6)

	a.Banner.Grid.AddItem(a.backBtn, 0, 0, 1, 1, 1, 5, false)
	a.Banner.Grid.AddItem(a.description, 0, 2, 2, 6, 1, 10, false)
	a.Banner.Grid.AddItem(a.paging, 3, 4, 1, 3, 1, 10, true)
	a.Banner.Grid.AddItem(a.list, 4, 0, 1, 8, 4, 10, false)
	return a
}

func (a *ArtistList) SetText(text string) {
	a.description.SetText(text)
}

// EnablePaging enables paging and shows page on banner
func (a *ArtistList) EnablePaging(enabled bool) {
	if a.pagingEnabled && enabled {
		return
	}
	if !a.pagingEnabled && !enabled {
		return
	}
	a.pagingEnabled = enabled
	if enabled {
		selectables := []twidgets.Selectable{a.backBtn, a.paging.Previous, a.paging.Next, a.list}
		a.Banner.Selectable = selectables
		a.Banner.Grid.AddItem(a.paging, 3, 4, 1, 3, 1, 10, true)
	} else {
		selectables := []twidgets.Selectable{a.backBtn, a.list}
		a.Banner.Selectable = selectables
		a.Banner.Grid.RemoveItem(a.paging)
		a.page.CurrentPage = 0
	}
}

func (a *ArtistList) Clear() {
	a.list.Clear()
	a.artists = make([]*ArtistCover, 0)
}

func (a *ArtistList) SetPage(paging interfaces.Paging) {
	a.paging.SetPage(paging.CurrentPage)
	a.paging.SetTotalPages(paging.TotalPages)
	a.page = paging
}

func (a *ArtistList) AddArtists(artists []*models.Artist) {
	items := make([]twidgets.ListItem, len(artists))

	offset := 0
	if a.pagingEnabled {
		offset = a.page.Offset()
	}

	for i, v := range artists {
		cover := newArtistCover(v)
		a.artists = append(a.artists, cover)
		if v.AlbumCount > 0 {
			cover.SetText(fmt.Sprintf("%d. %s\n%d albums %s",
				offset+i+1, v.Name, v.AlbumCount, util.SecToString(v.TotalDuration)))
		} else {
			cover.SetText(fmt.Sprintf("%d. %s\n %s",
				offset+i+1, v.Name, util.SecToString(v.TotalDuration)))
		}
		items[i] = cover
	}

	a.list.AddItems(items...)
}

func (a *ArtistList) selectArtist(index int) {
	if a.selectFunc != nil {
		artist := a.artists[index]
		a.selectFunc(artist.artist)
	}
}

func (a *ArtistList) selectPage(n int) {
	if a.selectPageFunc != nil {
		a.paging.SetPage(n)
		a.page.CurrentPage = n
		a.selectPageFunc(a.page)
	}
}

func (a *ArtistList) listHandler(key *tcell.EventKey) *tcell.EventKey {
	if key.Key() == tcell.KeyEnter && a.selectFunc != nil {
		index := a.list.GetSelectedIndex()
		artist := a.artists[index]
		a.selectFunc(artist.artist)
		return nil
	}
	return key
}

type ArtistCover struct {
	*cview.TextView
	artist *models.Artist
}

func newArtistCover(artist *models.Artist) *ArtistCover {
	a := &ArtistCover{
		TextView: cview.NewTextView(),
		artist:   artist,
	}
	a.SetBackgroundColor(config.Color.Background)
	a.SetTextColor(config.Color.Text)

	a.SetText(artist.Name)

	return a
}

func (a *ArtistCover) SetSelected(s twidgets.Selection) {
	if s == twidgets.Selected {
		a.SetTextColor(config.Color.TextSelected)
		a.SetBackgroundColor(config.Color.BackgroundSelected)
	} else if s == twidgets.Deselected {
		a.SetTextColor(config.Color.Text)
		a.SetBackgroundColor(config.Color.Background)
	} else if s == twidgets.Blurred {
		a.SetBackgroundColor(config.Color.TextDisabled)
	}
}
