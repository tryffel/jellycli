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
)

type MediaSelect int

const (
	MediaLatestMusic MediaSelect = iota
	MediaRecent
	MediaArtists
	MediaAlbumArtists
	MediaAlbums
	MediaSongs
	MediaPlaylists
	MediaFavoriteArtists
	MediaFavoriteAlbums
	MediaGenres
)

var mediaSelections = map[MediaSelect]string{
	MediaLatestMusic:     "Latest Music",
	MediaRecent:          "Recently played",
	MediaArtists:         "Artists",
	MediaAlbumArtists:    "Album Artists",
	MediaAlbums:          "Albums",
	MediaSongs:           "Songs",
	MediaPlaylists:       "Playlists",
	MediaFavoriteArtists: "Favorite Artists",
	MediaFavoriteAlbums:  "Favorite Albums",
	MediaGenres:          "Genres",
}

//MediaNavigation provides access to artists, albums, playlists
type MediaNavigation struct {
	*cview.Table
	selectFunc func(MediaSelect)
}

//NewMediaNavigation constructs new mediaNavigation. SelectFunc is called every time user
// wants to access given resource. SelectFunc can be nil.
func NewMediaNavigation(selectFunc func(selection MediaSelect)) *MediaNavigation {
	m := &MediaNavigation{
		Table:      cview.NewTable(),
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
	notImplemented := []MediaSelect{}

	for _, v := range notImplemented {
		cell := m.Table.GetCell(int(v), 0)
		cell.SetTextColor(config.Color.TextDisabled)
	}
}

func (m *MediaNavigation) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		key := event.Key()

		if key == tcell.KeyEnter && m.selectFunc != nil {
			index, _ := m.Table.GetSelection()
			m.selectFunc(MediaSelect(index))
		} else {
			m.Table.InputHandler()(event, setFocus)
		}
	}
}

// MouseHandler returns the mouse handler for this primitive.
func (m *MediaNavigation) MouseHandler() func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
	return m.WrapMouseHandler(func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
		// Pass events to context menu.
		if !m.InRect(event.Position()) {
			return false, nil
		}

		// Process mouse event.
		switch action {
		case cview.MouseLeftClick:
			setFocus(m)
			index := 0
			x, y := event.Position()
			rectX, rectY, width, height := m.GetInnerRect()
			if x < rectX || x >= rectX+width || y < rectY || y >= rectY+height {
				index = -1
			}

			index = y - rectY
			offset, _ := m.GetOffset()
			index += offset

			if index >= m.GetRowCount() {
				index = -1
			}

			if index >= 0 && index < m.GetRowCount() && m.selectFunc != nil {
				m.Table.Select(index, 0)
				m.selectFunc(MediaSelect(index))
			}
		}
		return
	})
}

func (m *MediaNavigation) SetCount(id MediaSelect, count int) {
	m.Table.SetCellSimple(int(id), 1, fmt.Sprint(count))
}

func tableCell(text string) *cview.TableCell {
	c := cview.NewTableCell(text)
	c.SetTextColor(config.Color.Text)
	c.SetAlign(cview.AlignLeft)
	return c
}
