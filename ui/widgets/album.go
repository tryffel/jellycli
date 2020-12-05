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
	"github.com/rivo/uniseg"
	"gitlab.com/tslocum/cview"
	"strings"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

type albumSong struct {
	*cview.TextView
	song        *models.Song
	showDiscNum bool
	index       int
	// is song being played now
	playing bool

	// allow overriding text input. If updateTextFunc != nil, use that to update, else use default album text format
	updateTextFunc func(a *albumSong)
}

func (a *albumSong) SetSelected(selected twidgets.Selection) {
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

func (a *albumSong) SetRect(x, y, w, h int) {
	_, _, ch, cw := a.GetRect()
	a.TextView.SetRect(x, y, w, h)
	if cw != w && a.song != nil {
		a.setText()
	}
	if ch != h {
	}
}

func (a *albumSong) setText() {
	if a.song == nil {
		return
	}
	if a.updateTextFunc != nil {
		a.updateTextFunc(a)
	} else {
		_, _, w, _ := a.GetRect()
		var name string
		if a.showDiscNum {
			name = fmt.Sprintf("%d %d. %s", a.song.DiscNumber, a.song.Index, a.song.Name)
		} else {
			name = fmt.Sprintf("%d. %s", a.index, a.song.Name)
		}

		text := a.getAlignedDuration(name)

		space := "      "
		artists := space

		// print artists if needed
		if len(a.song.Artists) > 1 {
			text += "\n"
			for i, v := range a.song.Artists {
				if i > 0 {
					artists += ", "
				}
				artists += v.Name
			}
			if len(artists) > w {
				artists = space + fmt.Sprintf("%d artists", len(a.song.Artists))
			} else {
				text += artists
			}
		} else if len(a.song.Artists) == 1 {
			if a.song.Artists[0].Id != a.song.AlbumArtist && a.song.AlbumArtist != "" {
				text += space + a.song.Artists[0].Name
			}
		}
		a.SetText(text)
	}
}

// add duration to text with space so that duration is aligned right
func (a *albumSong) getAlignedDuration(text string) string {
	_, _, w, _ := a.GetRect()
	nameLen := uniseg.GraphemeClusterCount(text)

	duration := util.SecToString(a.song.Duration)
	durationLen := len(duration)
	// width - duration - name - padding
	spaces := w - durationLen - nameLen - 2
	space := ""

	// calculate space needed between name etc and duration
	if spaces <= 0 {
		lines := cview.WordWrap(a.song.Name, w-2)
		if len(lines) >= 1 {
			text = lines[0] + "… "
		}
	} else {
		// add space as needed
		for {
			if len(space) == spaces {
				break
			}
			if len(space) < spaces-10 {
				space += "          "
			} else if len(space) < spaces-5 {
				space += "     "
			} else if len(space) < spaces-3 {
				space += "   "
			} else {
				space += " "
			}
		}
	}
	out := text + space + duration
	return out
}

func (a *albumSong) SetPlaying(playing bool) {
	a.playing = playing
}

// showDiscNum: whether to print disc number.
// overrideIndex: set -1 to use song index, else overrides index
func newAlbumSong(s *models.Song, showDiscNum bool, overrideIndex int) *albumSong {
	song := &albumSong{
		TextView:    cview.NewTextView(),
		song:        s,
		showDiscNum: showDiscNum,
		playing:     false,
	}

	if overrideIndex == -1 {
		song.index = s.Index
	} else {
		song.index = overrideIndex
	}

	song.SetBackgroundColor(config.Color.Background)
	song.SetTextColor(config.Color.Text)
	song.setText()
	song.SetBorderPadding(0, 0, 1, 1)

	return song
}

// AlbumView shows user a header (album name, info, buttons) and list of songs
type AlbumView struct {
	*itemList
	songs  []*albumSong
	artist *models.Artist
	album  *models.Album

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)

	similarBtn *button
	playBtn    *button
	dropDown   *dropDown

	similarFunc func(album *models.Album)
	context     contextOperator
}

//NewAlbumView initializes new album view
func NewAlbumview(playSong func(song *models.Song),
	playSongs func(songs []*models.Song), operator contextOperator) *AlbumView {
	a := &AlbumView{
		playSongFunc:  playSong,
		playSongsFunc: playSongs,

		similarBtn: newButton("Similar"),
		playBtn:    newButton("Play all"),
		context:    operator,
		dropDown:   newDropDown("Options"),
	}

	a.itemList = newItemList(a.playSong)
	a.list.ItemHeight = 2
	a.list.Padding = 1
	a.list.Grid.SetColumns(1, -1)

	a.reduceEnabled = true
	a.setReducerVisible = a.showReduceInput

	a.SetBorder(true)
	a.playBtn.SetSelectedFunc(a.playAlbum)

	a.Banner.Grid.SetRows(1, 1, 1, 1, -1, 3)
	a.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Banner.Grid.SetMinSize(1, 6)

	a.Banner.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Banner.Grid.AddItem(a.description, 0, 2, 2, 6, 1, 10, false)
	a.Banner.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, true)
	a.Banner.Grid.AddItem(a.dropDown, 3, 4, 1, 1, 1, 10, false)
	a.Banner.Grid.AddItem(a.list, 4, 0, 4, 8, 4, 10, false)

	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.dropDown, a.list}
	a.similarBtn.SetSelectedFunc(a.showSimilar)
	a.Banner.Selectable = selectables

	if a.context != nil {
		a.list.AddContextItem("Play all from here", 0, func(index int) {
			a.playFromSelected()
		})
		a.list.AddContextItem("View artist", 0, func(index int) {
			if index < len(a.songs) && a.context != nil {
				song := a.songs[0]
				a.context.ViewSongArtist(song.song)
			}
		})
		a.list.AddContextItem("Instant mix", 0, func(index int) {
			if index < len(a.songs) && a.context != nil {
				song := a.songs[0]
				a.context.InstantMix(song.song)
			}
		})
	}

	if a.context != nil {
		a.dropDown.AddOption("Instant mix", func() {
			a.context.InstantMix(a.artist)
		})
		a.dropDown.AddOption("View similar", func() {
			a.showSimilar()
		})
		a.dropDown.AddOption("Open in browser", func() {
			a.context.OpenInBrowser(a.album)
		})
	}

	a.itemList.initContextMenuList()
	return a
}

func (a *AlbumView) SetAlbum(album *models.Album, songs []*models.Song) {
	a.list.Clear()
	a.resetReduce()
	a.songs = make([]*albumSong, len(songs))
	items := make([]twidgets.ListItem, len(songs))

	itemTexts := make([]string, len(songs))

	album.SongCount = len(a.songs)
	a.album = album

	text := ""
	if album.Favorite {
		text += charFavorite + " "
	}

	text += album.Name
	if len(a.album.AdditionalArtists) > 1 {
		text += " ("
		for i, v := range a.album.AdditionalArtists {
			if i > 6 {
				remaining := len(a.album.AdditionalArtists) - i
				text += fmt.Sprintf(" and %d other artists", remaining)
				break
			}
			if i > 0 {
				text += ", "
			}
			text += v.Name
		}
		text += ")"

	} else if len(a.album.AdditionalArtists) == 1 {
		text += ": " + album.AdditionalArtists[0].Name
		text += " "
	}

	text += fmt.Sprintf("\n%d tracks  %s  %d",
		album.SongCount, util.SecToStringApproximate(album.Duration), album.Year)

	a.description.SetText(text)

	discs := map[int]bool{}
	for _, v := range songs {
		discs[v.DiscNumber] = true
	}
	album.DiscCount = len(discs)
	showDiscNum := album.DiscCount != 1
	for i, v := range songs {
		a.songs[i] = newAlbumSong(v, showDiscNum, -1)
		items[i] = a.songs[i]
		itemTexts[i] = strings.ToLower(v.Name)
	}

	a.list.AddItems(items...)
	a.items = items
	a.itemsTexts = itemTexts
	a.searchItemsSet()
}

func (a *AlbumView) SetArtist(artist *models.Artist) {
	a.artist = artist
}

func (a *AlbumView) playSong(index int) {
	if a.playSongFunc != nil {
		song := a.songs[index].song
		a.playSongFunc(song)
	}
}

func (a *AlbumView) playAlbum() {
	if a.playSongsFunc != nil {
		songs := make([]*models.Song, len(a.songs))
		for i, v := range a.songs {
			songs[i] = v.song
		}
		a.playSongsFunc(songs)
	}
}

func (a *AlbumView) playFromSelected() {
	if a.playSongsFunc != nil {
		index := a.getSelectedIndex()
		songs := make([]*models.Song, len(a.songs)-index)
		for i, v := range a.songs[index:] {
			songs[i] = v.song
		}
		a.playSongsFunc(songs)
	}
}

func (a *AlbumView) showSimilar() {
	if a.similarFunc != nil {
		a.similarFunc(a.album)
	}
}

func (a *AlbumView) showReduceInput(visible bool) {
	if visible {
		a.Grid.AddItem(a.reduceInput, 5, 0, 1, 10, 1, 20, false)
		a.Grid.RemoveItem(a.list)
		a.Grid.AddItem(a.list, 4, 0, 1, 10, 6, 20, false)
	} else {
		a.Grid.RemoveItem(a.reduceInput)
		a.Grid.RemoveItem(a.list)
		a.Grid.AddItem(a.list, 4, 0, 2, 10, 6, 20, false)
	}

}
