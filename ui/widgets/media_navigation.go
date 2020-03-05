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
)

type MediaSelect int

const (
	MediaLatestMusic MediaSelect = iota
	MediaRecent
	MediaArtists
	MediaAlbums
	MediaSongs
	MediaPlaylists
	MediaFavoriteArtists
	MediaFavoriteAlbums
)

var mediaSelections = map[MediaSelect]string{
	MediaLatestMusic:     "Latest Music",
	MediaRecent:          "Recently played",
	MediaArtists:         "Artists",
	MediaAlbums:          "Albums",
	MediaSongs:           "Songs",
	MediaPlaylists:       "Playlists",
	MediaFavoriteArtists: "Favorite Artists",
	MediaFavoriteAlbums:  "Favorite Albums",
}

//MediaNavigation provides access to artists, albums, playlists
type MediaNavigation struct {
	*tview.Table
	selectFunc func(MediaSelect)
}

//NewMediaNavigation constructs new mediaNavigation. SelectFunc is called every time user
// wants to access given resource. SelectFunc can be nil.
func NewMediaNavigation(selectFunc func(selection MediaSelect)) *MediaNavigation {
	m := &MediaNavigation{
		Table:      tview.NewTable(),
		selectFunc: selectFunc,
	}

	m.SetBorder(true)
	m.SetBorderColor(config.Color.Border)
	m.SetBackgroundColor(config.Color.NavBar.Background)
	m.SetBorder(true)
	m.SetSelectable(true, false)
	m.SetSelectedStyle(config.Color.TextSelected, config.Color.BackgroundSelected, 0)

	for i, v := range mediaSelections {
		cell := tableCell(v)
		m.Table.SetCell(int(i), 0, cell)
	}

	m.markDisabledMethods()
	return m
}

func (m *MediaNavigation) markDisabledMethods() {
	// colorize methods that are not implemented
	notImplemented := []MediaSelect{
		MediaRecent,
		MediaArtists,
		MediaAlbums,
		MediaFavoriteAlbums,
	}

	for _, v := range notImplemented {
		cell := m.Table.GetCell(int(v), 0)
		cell.SetTextColor(config.Color.TextDisabled)
	}
}

func (m *MediaNavigation) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()

		if key == tcell.KeyEnter && m.selectFunc != nil {
			index, _ := m.Table.GetSelection()
			m.selectFunc(MediaSelect(index))
		} else {
			m.Table.InputHandler()(event, setFocus)
		}
	}
}

func (m *MediaNavigation) SetCount(id MediaSelect, count int) {
	m.Table.SetCellSimple(int(id), 1, fmt.Sprint(count))
}

func tableCell(text string) *tview.TableCell {
	c := tview.NewTableCell(text)
	c.SetTextColor(config.Color.Text)
	c.SetAlign(tview.AlignLeft)
	return c
}
