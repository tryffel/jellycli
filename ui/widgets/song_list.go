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
	"strings"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/twidgets"
)

// SongList shows a list of songs and optional paging
type SongList struct {
	*itemList
	paging *PageSelector
	songs  []*albumSong
	title  string

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)
	showPage      func(paging interfaces.Paging)

	playBtn *button
	context contextOperator
	page    interfaces.Paging
}

// NewSongList initializes new song list
func NewSongList(playSong func(song *models.Song), playSongs func(songs []*models.Song),
	operator contextOperator) *SongList {
	p := &SongList{
		playSongFunc:  playSong,
		playSongsFunc: playSongs,
		context:       operator,
		playBtn:       newButton("Play all"),
	}

	p.itemList = newItemList(p.selectSong)
	p.paging = NewPageSelector(p.selectPage)

	p.list.ItemHeight = 2
	p.list.Padding = 1
	p.list.Grid.SetColumns(1, -1)

	p.playBtn.SetSelectedFunc(p.playAll)
	p.Banner.Grid.SetRows(1, 1, 1, 1, -1, 3)
	p.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	p.Banner.Grid.SetMinSize(1, 6)

	p.Banner.Grid.AddItem(p.prevBtn, 0, 0, 1, 1, 1, 5, false)
	p.Banner.Grid.AddItem(p.description, 0, 2, 2, 6, 1, 10, false)
	p.Banner.Grid.AddItem(p.playBtn, 3, 2, 1, 1, 1, 10, true)
	p.Banner.Grid.AddItem(p.paging, 3, 4, 1, 3, 1, 10, true)
	p.Banner.Grid.AddItem(p.list, 4, 0, 2, 8, 4, 10, false)

	selectables := []twidgets.Selectable{p.prevBtn, p.playBtn, p.paging.Previous, p.paging.Next, p.list}
	p.Banner.Selectable = selectables
	p.title = "All songs"

	if p.context != nil {
		p.list.AddContextItem("View album", 0, func(index int) {
			selected := p.getSelectedIndex()
			song := p.songs[selected]
			p.context.ViewSongAlbum(song.song)
		})
		p.list.AddContextItem("View artist", 0, func(index int) {
			selected := p.getSelectedIndex()
			song := p.songs[selected]
			p.context.ViewSongArtist(song.song)
		})
		p.list.AddContextItem("Instant mix", 0, func(index int) {
			selected := p.getSelectedIndex()
			song := p.songs[selected]
			p.context.InstantMix(song.song)
		})

	}

	p.reduceEnabled = true
	p.setReducerVisible = p.showReduceInput

	p.itemList.initContextMenuList()
	return p
}

func (s *SongList) setTitle(title string) {
	s.title = title
}

func (s *SongList) SetSongs(songs []*models.Song, page interfaces.Paging) {
	s.list.Clear()
	s.resetReduce()
	s.page = page
	s.songs = make([]*albumSong, len(songs))
	items := make([]twidgets.ListItem, len(songs))
	itemTexts := make([]string, len(songs))

	text := fmt.Sprintf("%s: %d songs", s.title, page.TotalItems)

	s.description.SetText(text)

	offset := page.CurrentPage * page.PageSize

	for i, v := range songs {
		s.songs[i] = newAlbumSong(v, false, offset+i+1)
		s.songs[i].updateTextFunc = s.updateSongText
		items[i] = s.songs[i]

		itemText := songs[i].Name
		if len(songs[i].Artists) > 0 {
			for _, v := range songs[i].Artists {
				itemText += " " + v.Name
			}
		}
		itemTexts[i] = strings.ToLower(itemText)
	}
	s.list.AddItems(items...)

	s.paging.SetPage(page.CurrentPage)
	s.paging.SetTotalPages(page.TotalPages)

	s.items = items
	s.itemsTexts = itemTexts
}

func (s *SongList) selectSong(index int) {
	if s.playSongFunc != nil {
		song := s.songs[index].song
		s.playSongFunc(song)
	}
}

func (s *SongList) playAll() {
	if s.playSongsFunc != nil {
		songs := make([]*models.Song, len(s.songs))
		for i, v := range s.songs {
			songs[i] = v.song
		}
		s.playSongsFunc(songs)
	}
}

func (s *SongList) updateSongText(song *albumSong) {
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

func (s *SongList) selectPage(n int) {
	s.paging.SetPage(n)

	s.page.CurrentPage = n
	if s.showPage != nil {
		s.showPage(s.page)
	}
	s.resetReduce()
}

func (s *SongList) showReduceInput(visible bool) {
	if visible {
		s.Banner.Grid.AddItem(s.reduceInput, 5, 0, 1, 10, 1, 10, false)
		s.Banner.Grid.RemoveItem(s.list)
		s.Banner.Grid.AddItem(s.list, 4, 0, 1, 10, 6, 20, false)
	} else {
		s.Banner.Grid.RemoveItem(s.reduceInput)
		s.Banner.Grid.RemoveItem(s.list)
		s.Banner.Grid.AddItem(s.list, 4, 0, 2, 10, 6, 20, false)
	}
}
