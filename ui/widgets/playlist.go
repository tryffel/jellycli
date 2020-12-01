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
	"strings"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

// AlbumView shows user a header (album name, info, buttons) and list of songs
type PlaylistView struct {
	*itemList
	songs    []*albumSong
	playlist *models.Playlist

	context contextOperator

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)

	playBtn *button
	options *dropDown
}

//NewAlbumView initializes new album view
func NewPlaylistView(playSong func(song *models.Song), playSongs func(songs []*models.Song),
	operator contextOperator) *PlaylistView {
	p := &PlaylistView{
		playSongFunc:  playSong,
		playSongsFunc: playSongs,

		playBtn: newButton("Play all"),
		context: operator,
		options: newDropDown("Options"),
	}

	p.itemList = newItemList(p.playSong)
	p.list.ItemHeight = 2
	p.list.Padding = 1
	p.list.SetInputCapture(p.listHandler)
	p.list.Grid.SetColumns(1, -1)

	p.playBtn.SetSelectedFunc(p.playAll)

	p.Banner.Grid.SetRows(1, 1, 1, 1, -1, 3)
	p.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	p.Banner.Grid.SetMinSize(1, 6)

	p.Banner.Grid.AddItem(p.prevBtn, 0, 0, 1, 1, 1, 5, false)
	p.Banner.Grid.AddItem(p.description, 0, 2, 2, 6, 1, 10, false)
	p.Banner.Grid.AddItem(p.playBtn, 3, 2, 1, 1, 1, 10, true)
	p.Banner.Grid.AddItem(p.options, 3, 4, 1, 1, 1, 10, false)
	p.Banner.Grid.AddItem(p.list, 4, 0, 2, 8, 4, 10, false)

	selectables := []twidgets.Selectable{p.prevBtn, p.playBtn, p.options, p.list}
	p.Banner.Selectable = selectables

	p.reduceEnabled = true
	p.setReducerVisible = p.showReduceInput

	if p.context != nil {
		p.list.AddContextItem("Play all from here", 0, func(index int) {
			p.playFromSelected()
		})
		p.list.AddContextItem("View album", 0, func(index int) {
			selected := p.getSelectedIndex()
			song := p.songs[selected]
			p.context.ViewSongAlbum(song.song)
		})
		p.list.AddContextItem("View artist", 0, func(index int) {
			if index < len(p.songs) && p.context != nil {
				index := p.getSelectedIndex()
				song := p.songs[index]
				p.context.ViewSongArtist(song.song)
			}
		})
		p.list.AddContextItem("Instant mix", 0, func(index int) {
			if index < len(p.songs) && p.context != nil {
				index := p.getSelectedIndex()
				song := p.songs[index]
				p.context.InstantMix(song.song)
			}
		})

		p.options.AddOption("Instant mix", func() {
			p.context.InstantMix(p.playlist)
		})

		p.options.AddOption("Open in browser", func() {
			p.context.OpenInBrowser(p.playlist)
		})
	}

	p.list.ContextMenuList().SetBorder(true)
	p.list.ContextMenuList().SetBackgroundColor(config.Color.Background)
	p.list.ContextMenuList().SetBorderColor(config.Color.BorderFocus)
	p.list.ContextMenuList().SetSelectedBackgroundColor(config.Color.BackgroundSelected)
	p.list.ContextMenuList().SetMainTextColor(config.Color.Text)
	p.list.ContextMenuList().SetSelectedTextColor(config.Color.TextSelected)

	return p
}

func (p *PlaylistView) SetPlaylist(playlist *models.Playlist) {
	p.list.Clear()
	p.resetReduce()
	p.playlist = playlist
	p.songs = make([]*albumSong, len(playlist.Songs))
	items := make([]twidgets.ListItem, len(playlist.Songs))

	text := playlist.Name

	text += fmt.Sprintf("\n%d tracks  %s",
		len(playlist.Songs), util.SecToStringApproximate(playlist.Duration))

	p.description.SetText(text)
	itemTexts := make([]string, len(playlist.Songs))

	for i, v := range playlist.Songs {
		p.songs[i] = newAlbumSong(v, false, i+1)
		p.songs[i].updateTextFunc = p.updateSongText
		items[i] = p.songs[i]

		itemText := v.Name
		for _, artist := range v.Artists {
			itemText += " " + artist.Name
		}
		itemTexts[i] = strings.ToLower(itemText)
	}

	p.list.AddItems(items...)
	p.items = items
	p.searchItemsSet()
	p.itemsTexts = itemTexts
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

func (p *PlaylistView) playFromSelected() {
	if p.playSongsFunc != nil {
		index := p.list.GetSelectedIndex()
		songs := make([]*models.Song, len(p.songs)-index)
		for i, v := range p.songs[index:] {
			songs[i] = v.song
		}
		p.playSongsFunc(songs)
	}
}

func (p *PlaylistView) listHandler(key *tcell.EventKey) *tcell.EventKey {
	if key.Key() == tcell.KeyEnter && key.Modifiers() == tcell.ModNone {
		index := p.list.GetSelectedIndex()
		p.playSong(index)
		return nil
	}
	return key
}

func (p *PlaylistView) updateSongText(song *albumSong) {
	var name string
	if song.showDiscNum {
		name = fmt.Sprintf("%d %d. %s", song.song.DiscNumber, song.song.Index, song.song.Name)
	} else {
		name = fmt.Sprintf("%d. %s", song.index, song.song.Name)
	}

	text := song.getAlignedDuration(name)
	if len(song.song.Artists) > 0 {
		text += "\n     " + song.song.Artists[0].Name

	}
	song.SetText(text)
}

func (p *PlaylistView) showReduceInput(visible bool) {
	if visible {
		p.Grid.AddItem(p.reduceInput, 5, 0, 1, 10, 1, 20, false)
		p.Grid.RemoveItem(p.list)
		p.Grid.AddItem(p.list, 4, 0, 1, 10, 6, 20, false)
	} else {
		p.Grid.RemoveItem(p.reduceInput)
		p.Grid.RemoveItem(p.list)
		p.Grid.AddItem(p.list, 4, 0, 2, 10, 6, 20, false)
	}
}
