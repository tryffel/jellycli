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
	"strings"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

//AlbumCover is a simple cover for album, it shows
// album name, year and possible artists
type PlaylistCover struct {
	*cview.TextView
	album   *models.Playlist
	index   int
	name    string
	year    int
	artists []string
}

func NewPlaylistCover(index int, playlist *models.Playlist) *PlaylistCover {
	a := &PlaylistCover{
		TextView: cview.NewTextView(),
		album:    playlist,
		index:    index,
	}

	a.SetBorder(false)
	a.SetBackgroundColor(config.Color.Background)
	a.SetBorderPadding(0, 0, 1, 1)
	a.SetTextColor(config.Color.Text)
	ar := printArtists(a.artists, 40)
	text := fmt.Sprintf("%d. %s\n%d songs, %s", index, playlist.Name,
		playlist.SongCount, util.SecToStringApproximate(playlist.Duration))
	if ar != "" {
		text += "\n" + ar
	}

	a.TextView.SetText(text)
	return a
}

func (a *PlaylistCover) SetRect(x, y, w, h int) {
	a.TextView.SetRect(x, y, w, h)
}

func (a *PlaylistCover) SetSelected(selected twidgets.Selection) {
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

//Playlists shows playlists
type Playlists struct {
	*itemList
	selectFunc     func(album *models.Playlist)
	playlistCovers []*PlaylistCover
	playBtn        *button
}

func (pl *Playlists) Clear() {
	pl.list.Clear()
	pl.playlistCovers = make([]*PlaylistCover, 0)
	pl.resetReduce()
}

// SetPlaylist sets albums
func (pl *Playlists) SetPlaylists(playlists []*models.Playlist) {
	pl.list.Clear()
	pl.playlistCovers = make([]*PlaylistCover, len(playlists))

	itemTexts := make([]string, len(playlists))

	items := make([]twidgets.ListItem, len(playlists))
	for i, v := range playlists {
		cover := NewPlaylistCover(i+1, v)
		items[i] = cover
		pl.playlistCovers[i] = cover

		itemTexts[i] = strings.ToLower(v.Name)
	}
	pl.list.AddItems(items...)
	pl.description.SetText(fmt.Sprintf("Playlists: %d", len(playlists)))
	pl.items = items
	pl.itemsTexts = itemTexts
	pl.searchItemsSet()
}

// NewPlaylists constructs new playlist view
func NewPlaylists(selectPlaylist func(playlist *models.Playlist)) *Playlists {
	a := &Playlists{
		selectFunc: selectPlaylist,
		playBtn:    newButton("Play all"),
	}
	a.itemList = newItemList(a.selectAlbum)
	a.itemList.list.ItemHeight = 3
	a.itemList.reduceEnabled = true
	a.itemList.setReducerVisible = a.showReduceInput
	a.list.Grid.SetColumns(-1, 5)

	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.list}
	a.prevBtn.SetSelectedFunc(a.goBack)
	a.Banner.Selectable = selectables
	a.Grid.SetRows(1, 1, 1, 1, -1, 3)
	a.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Grid.SetMinSize(1, 6)
	a.Grid.SetBackgroundColor(config.Color.Background)
	a.description.SetText("Playlists")
	a.list.Grid.SetColumns(1, -1)
	a.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Grid.AddItem(a.description, 0, 2, 2, 6, 1, 10, false)
	a.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.list, 4, 0, 2, 8, 6, 20, false)

	a.listFocused = false
	return a
}

func (pl *Playlists) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		pl.Banner.InputHandler()(event, setFocus)
	}
}

func (pl *Playlists) selectAlbum(index int) {
	if pl.selectFunc != nil {
		album := pl.playlistCovers[index]
		pl.selectFunc(album.album)
		pl.resetReduce()
	}
}

func (pl *Playlists) showReduceInput(visible bool) {
	if visible {
		pl.Grid.AddItem(pl.reduceInput, 5, 0, 1, 10, 1, 20, false)
		pl.Grid.RemoveItem(pl.list)
		pl.Grid.AddItem(pl.list, 4, 0, 1, 10, 6, 20, false)
	} else {
		pl.Grid.RemoveItem(pl.reduceInput)
		pl.Grid.RemoveItem(pl.list)
		pl.Grid.AddItem(pl.list, 4, 0, 2, 10, 6, 20, false)
	}
}
