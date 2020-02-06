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
	"tryffel.net/go/twidgets"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/models"
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

	name        *tview.TextView
	description *tview.TextView

	prevBtn    *tview.Button
	infoBtn    *tview.Button
	anotherBtn *tview.Button
	similarBtn *tview.Button
	prevFunc   func()
}

func NewArtistHeader(prevFunc func()) *ArtistHeader {
	a := &ArtistHeader{
		Grid:        tview.NewGrid(),
		artist:      &models.Artist{},
		name:        tview.NewTextView(),
		description: tview.NewTextView(),
		prevBtn:     tview.NewButton("Back"),
		prevFunc:    prevFunc,
		anotherBtn:  tview.NewButton("Play all"),
		similarBtn:  tview.NewButton("Similar"),
	}

	a.name.SetBorderPadding(0, 0, 1, 1)
	a.description.SetBorderPadding(0, 0, 1, 1)
	a.name.SetText(a.artist.Name)
	a.description.SetText(fmt.Sprintf("Albums: %d, Total: %s",
		a.artist.AlbumCount, SecToStringApproximate(a.artist.TotalDuration)))
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

	a.Grid.AddItem(a.prevBtn, 1, 1, 1, 1, 1, 5, true)
	a.Grid.AddItem(a.name, 1, 3, 1, 5, 1, 10, false)
	a.Grid.AddItem(a.description, 2, 3, 1, 5, 1, 10, false)
	a.Grid.AddItem(a.anotherBtn, 4, 3, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.similarBtn, 4, 5, 1, 1, 1, 10, false)
	return a
}

func (a *ArtistHeader) SetArtist(artist *models.Artist) {
	a.artist = artist
	a.name.SetText(a.artist.Name)
	a.description.SetText(fmt.Sprintf("Albums: %d, Total: %s",
		a.artist.AlbumCount, SecToStringApproximate(a.artist.TotalDuration)))
}

//AlbumCover is a simple cover for album, it shows
// album name, year and possible artists
type AlbumCover struct {
	*tview.TextView
	name    string
	year    int
	artists []string
}

func NewAlbumCover(name string, year int, additionalArtists []string) *AlbumCover {
	a := &AlbumCover{
		TextView: tview.NewTextView(),
		name:     name,
		year:     year,
		artists:  additionalArtists,
	}

	a.SetBorder(true)
	a.SetBorderPadding(0, 0, 1, 1)
	ar := printArtists(a.artists, 40)
	text := fmt.Sprintf("%s\n%d", name, year)
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
		a.SetBorderColor(tcell.ColorBlue)
		a.SetBorderAttributes(tcell.AttrBold)
	} else {
		a.SetBorderColor(tcell.ColorGray)
		a.SetBorderAttributes(tcell.AttrNone)
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
	//list *AlbumList
	list   *twidgets.ScrollList
	header *ArtistHeader
}

func (a *ArtistView) AddAlbum(c *AlbumCover) {
	a.list.AddItem(c)
}

func (a *ArtistView) Clear() {
	a.list.Clear()
	a.header.SetArtist(nil)
}

func (a *ArtistView) SetArtist(artist *models.Artist) {
	a.header.SetArtist(artist)
}

//NewArtistView constructs new artist view
func NewArtistView() *ArtistView {
	a := &ArtistView{
		Grid:   tview.NewGrid(),
		list:   twidgets.NewScrollList(nil),
		header: NewArtistHeader(nil),
	}
	a.list.ItemHeight = 5

	a.Grid.SetRows(5, -1)
	a.Grid.SetColumns(-1)

	a.Grid.AddItem(a.header, 0, 0, 1, 1, 6, 25, false)
	a.Grid.AddItem(a.list, 1, 0, 1, 1, 6, 25, true)

	return a
}
