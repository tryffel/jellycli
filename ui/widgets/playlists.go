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

func (a *Playlists) AddAlbum(c *PlaylistCover) {
	a.list.AddItem(c)
	a.playlistCovers = append(a.playlistCovers, c)
}

func (a *Playlists) Clear() {
	a.list.Clear()
	a.playlistCovers = make([]*PlaylistCover, 0)
}

// SetPlaylist sets albums
func (a *Playlists) SetPlaylists(playlists []*models.Playlist) {
	a.list.Clear()
	a.playlistCovers = make([]*PlaylistCover, len(playlists))

	items := make([]twidgets.ListItem, len(playlists))
	for i, v := range playlists {
		cover := NewPlaylistCover(i+1, v)
		items[i] = cover
		a.playlistCovers[i] = cover
	}
	a.list.AddItems(items...)
	a.description.SetText(fmt.Sprintf("Playlists: %d", len(playlists)))
}

// NewPlaylists constructs new playlist view
func NewPlaylists(selectPlaylist func(playlist *models.Playlist)) *Playlists {
	a := &Playlists{

		selectFunc: selectPlaylist,
		playBtn:    newButton("Play all"),
	}
	a.itemList = newItemList(a.selectAlbum)
	a.itemList.list.ItemHeight = 3
	a.list.Grid.SetColumns(-1, 5)

	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.list}
	a.prevBtn.SetSelectedFunc(a.goBack)
	a.Banner.Selectable = selectables
	a.Grid.SetRows(1, 1, 1, 1, -1)
	a.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Grid.SetMinSize(1, 6)
	a.Grid.SetBackgroundColor(config.Color.Background)
	a.description.SetText("Playlists")
	a.list.Grid.SetColumns(1, -1)
	a.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Grid.AddItem(a.description, 0, 2, 2, 6, 1, 10, false)
	a.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, false)
	a.Grid.AddItem(a.list, 4, 0, 1, 8, 6, 20, false)

	a.listFocused = false
	return a
}

func (a *Playlists) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		a.Banner.InputHandler()(event, setFocus)
	}
}

func (a *Playlists) selectAlbum(index int) {
	if a.selectFunc != nil {
		index := a.list.GetSelectedIndex()
		album := a.playlistCovers[index]
		a.selectFunc(album.album)
	}
}
