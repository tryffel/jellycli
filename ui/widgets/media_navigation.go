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
	"tryffel.net/pkg/jellycli/config"
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
	MediaFavoritAlbums
)

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
	m.SetBorderColor(config.ColorBorder)
	m.SetBackgroundColor(config.ColorBackground)
	m.SetBorder(true)
	m.SetBorderColor(config.ColorBorder)
	m.SetSelectable(true, false)
	m.SetSelectedStyle(config.ColorPrimary, config.ColorBorder, 0)

	type keyValue struct {
		name  string
		count int
	}

	items := []keyValue{
		{"Latest Music", -1},
		{"Recently played", -1},
		{"Artists", 10},
		{"Albums", 20},
		{"Songs", 62},
		{"Playlists", 3},
		{"Favorite Artists", 4},
		{"Favorite Albums", 6},
	}

	for i, v := range items {
		cell := tableCell(v.name)
		m.Table.SetCell(i, 0, cell)
		if v.count > -1 {
			m.Table.SetCellSimple(i, 1, fmt.Sprint(v.count))
			cell = tableCell(fmt.Sprint(v.count))
			m.Table.SetCell(i, 1, cell)
		}
	}
	return m
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
