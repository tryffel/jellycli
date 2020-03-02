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
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

type ArtistList struct {
	*twidgets.ScrollList
	*previous
	selectFunc func(artist *models.Artist)
	artists    []*ArtistCover
}

func NewArtistList(selectFunc func(artist *models.Artist)) *ArtistList {
	a := &ArtistList{
		selectFunc: selectFunc,
		artists:    make([]*ArtistCover, 0),
		previous:   &previous{},
	}
	a.ScrollList = twidgets.NewScrollList(a.selectArtist)

	a.Padding = 1
	a.ItemHeight = 2

	a.SetBorder(true)
	a.SetBorderColor(config.Color.Border)
	a.SetBackgroundColor(config.Color.Background)
	a.SetBorder(true)
	return a
}

func (a *ArtistList) Clear() {
	a.ScrollList.Clear()
	a.artists = make([]*ArtistCover, 0)
}

func (a *ArtistList) AddArtists(artists []*models.Artist) {
	for i, v := range artists {
		cover := newArtistCover(v)
		a.artists = append(a.artists, cover)
		a.AddItem(cover)

		if v.AlbumCount > 0 {
			cover.SetText(fmt.Sprintf("%d. %s\n%d albums %s",
				i+1, v.Name, v.AlbumCount, util.SecToString(v.TotalDuration)))
		} else {
			cover.SetText(fmt.Sprintf("%d. %s\n %s",
				i+1, v.Name, util.SecToString(v.TotalDuration)))
		}
	}
}

func (a *ArtistList) selectArtist(index int) {
	if a.selectFunc != nil {
		artist := a.artists[index]
		a.selectFunc(artist.artist)
	}
}

type ArtistCover struct {
	*tview.TextView
	artist *models.Artist
}

func newArtistCover(artist *models.Artist) *ArtistCover {
	a := &ArtistCover{
		TextView: tview.NewTextView(),
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

	}
}
