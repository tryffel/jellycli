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
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

// AlbumView shows user a header (album name, info, buttons) and list of songs
type PlaylistView struct {
	*twidgets.Banner
	*previous
	list        *twidgets.ScrollList
	songs       []*albumSong
	playlist    *models.Playlist
	listFocused bool

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)

	description *tview.TextView
	prevBtn     *button
	playBtn     *button
	prevFunc    func()
}

//NewAlbumView initializes new album view
func NewPlaylistView(playSong func(song *models.Song), playSongs func(songs []*models.Song)) *PlaylistView {
	p := &PlaylistView{
		Banner:        twidgets.NewBanner(),
		previous:      &previous{},
		list:          twidgets.NewScrollList(nil),
		playSongFunc:  playSong,
		playSongsFunc: playSongs,

		description: tview.NewTextView(),
		prevBtn:     newButton("Back"),
		playBtn:     newButton("Play all"),
	}

	p.list.ItemHeight = 2
	p.list.Padding = 1
	p.list.SetInputCapture(p.listHandler)
	p.list.SetBorder(true)
	p.list.SetBorderColor(config.Color.Border)
	p.list.Grid.SetColumns(1, -1)

	p.SetBorder(true)
	p.SetBorderColor(config.Color.Border)
	p.list.SetBackgroundColor(config.Color.Background)
	p.Grid.SetBackgroundColor(config.Color.Background)
	p.listFocused = false
	p.playBtn.SetSelectedFunc(p.playAll)

	p.Banner.Grid.SetRows(1, 1, 1, 1, -1)
	p.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	p.Banner.Grid.SetMinSize(1, 6)

	p.Banner.Grid.AddItem(p.prevBtn, 0, 0, 1, 1, 1, 5, false)
	p.Banner.Grid.AddItem(p.description, 0, 2, 2, 6, 1, 10, false)
	p.Banner.Grid.AddItem(p.playBtn, 3, 2, 1, 1, 1, 10, true)
	p.Banner.Grid.AddItem(p.list, 4, 0, 1, 8, 4, 10, false)

	btns := []*button{p.prevBtn, p.playBtn}
	selectables := []twidgets.Selectable{p.prevBtn, p.playBtn, p.list}
	for _, btn := range btns {
		btn.SetLabelColor(config.Color.ButtonLabel)
		btn.SetLabelColorActivated(config.Color.ButtonLabelSelected)
		btn.SetBackgroundColor(config.Color.ButtonBackground)
		btn.SetBackgroundColorActivated(config.Color.ButtonBackgroundSelected)
	}

	p.prevBtn.SetSelectedFunc(p.goBack)

	p.Banner.Selectable = selectables
	p.description.SetBackgroundColor(config.Color.Background)
	p.description.SetTextColor(config.Color.Text)
	return p
}

func (p *PlaylistView) SetPlaylist(playlist *models.Playlist) {
	p.list.Clear()
	p.playlist = playlist
	p.songs = make([]*albumSong, len(playlist.Songs))
	items := make([]twidgets.ListItem, len(playlist.Songs))

	text := playlist.Name

	text += fmt.Sprintf("\n%d tracks  %s",
		len(playlist.Songs), util.SecToStringApproximate(playlist.Duration))

	p.description.SetText(text)

	for i, v := range playlist.Songs {
		p.songs[i] = newAlbumSong(v, false, i+1)
		items[i] = p.songs[i]
	}

	p.list.AddItems(items...)
}

func (p *PlaylistView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if p.listFocused {
			index := p.list.GetSelectedIndex()
			if index == 0 && (key == tcell.KeyUp || key == tcell.KeyCtrlK) {
				p.listFocused = false
				p.prevBtn.Focus(func(p tview.Primitive) {})
				p.list.Blur()
			} else if key == tcell.KeyEnter {
				p.playSong(index)
			} else {
				p.list.InputHandler()(event, setFocus)
			}
		} else {
			if key == tcell.KeyDown || key == tcell.KeyCtrlJ {
				p.listFocused = true
				p.list.Focus(func(p tview.Primitive) {})
			} else {
			}
		}
	}
}

func (p *PlaylistView) playSong(index int) {
	if p.playSongFunc != nil {
		song := p.songs[index].song
		p.playSongFunc(song)
	}
}

func (p *PlaylistView) playAll() {
	if p.playSongsFunc != nil {
		songs := make([]*models.Song, len(p.songs))
		for i, v := range p.songs {
			songs[i] = v.song
		}
		p.playSongsFunc(songs)
	}
}

func (p *PlaylistView) listHandler(key *tcell.EventKey) *tcell.EventKey {
	if key.Key() == tcell.KeyEnter {
		index := p.list.GetSelectedIndex()
		p.playSong(index)
		return nil
	}
	return key
}
