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
	"github.com/rivo/uniseg"
	"gitlab.com/tslocum/cview"
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
	*twidgets.Banner
	*previous
	list        *twidgets.ScrollList
	songs       []*albumSong
	artist      *models.Artist
	album       *models.Album
	listFocused bool

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)

	description *cview.TextView
	prevBtn     *button
	similarBtn  *button
	playBtn     *button
	dropDown    *dropDown

	prevFunc    func()
	similarFunc func(album *models.Album)
	context     contextOperator
}

//NewAlbumView initializes new album view
func NewAlbumview(playSong func(song *models.Song),
	playSongs func(songs []*models.Song), operator contextOperator) *AlbumView {
	a := &AlbumView{
		Banner:        twidgets.NewBanner(),
		previous:      &previous{},
		list:          twidgets.NewScrollList(nil),
		playSongFunc:  playSong,
		playSongsFunc: playSongs,

		description: cview.NewTextView(),
		prevBtn:     newButton("Back"),
		similarBtn:  newButton("Similar"),
		playBtn:     newButton("Play all"),
		context:     operator,
		dropDown:    newDropDown("Options"),
	}

	a.list.ItemHeight = 2
	a.list.Padding = 1
	a.list.SetInputCapture(a.listHandler)
	a.list.SetBorder(true)
	a.list.SetBorderColor(config.Color.Border)
	a.list.Grid.SetColumns(1, -1)

	a.SetBorder(true)
	a.SetBorderColor(config.Color.Border)
	a.list.SetBackgroundColor(config.Color.Background)
	a.Grid.SetBackgroundColor(config.Color.Background)
	a.listFocused = false
	a.playBtn.SetSelectedFunc(a.playAlbum)

	a.Banner.Grid.SetRows(1, 1, 1, 1, -1)
	a.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	a.Banner.Grid.SetMinSize(1, 6)

	a.Banner.Grid.AddItem(a.prevBtn, 0, 0, 1, 1, 1, 5, false)
	a.Banner.Grid.AddItem(a.description, 0, 2, 2, 6, 1, 10, false)
	a.Banner.Grid.AddItem(a.playBtn, 3, 2, 1, 1, 1, 10, true)
	a.Banner.Grid.AddItem(a.dropDown, 3, 4, 1, 1, 1, 10, false)
	a.Banner.Grid.AddItem(a.list, 4, 0, 1, 8, 4, 10, false)

	btns := []*button{a.prevBtn, a.playBtn, a.similarBtn}
	selectables := []twidgets.Selectable{a.prevBtn, a.playBtn, a.dropDown, a.list}
	for _, btn := range btns {
		btn.SetLabelColor(config.Color.ButtonLabel)
		btn.SetLabelColorActivated(config.Color.ButtonLabelSelected)
		btn.SetBackgroundColor(config.Color.ButtonBackground)
		btn.SetBackgroundColorActivated(config.Color.ButtonBackgroundSelected)
	}

	a.prevBtn.SetSelectedFunc(a.goBack)
	a.similarBtn.SetSelectedFunc(a.showSimilar)

	a.Banner.Selectable = selectables
	a.description.SetBackgroundColor(config.Color.Background)
	a.description.SetTextColor(config.Color.Text)

	if a.context != nil {
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
	}
	if a.context != nil {
		a.dropDown.AddOption("View similar", func() {
			a.showSimilar()
		})
		a.dropDown.AddOption("View in browser", func() {
			a.context.OpenInBrowser(a.album)
		})
	}

	a.list.ContextMenuList().SetBorder(true)
	a.list.ContextMenuList().SetBackgroundColor(config.Color.Background)
	a.list.ContextMenuList().SetBorderColor(config.Color.BorderFocus)
	a.list.ContextMenuList().SetSelectedBackgroundColor(config.Color.BackgroundSelected)
	a.list.ContextMenuList().SetMainTextColor(config.Color.Text)
	a.list.ContextMenuList().SetSelectedTextColor(config.Color.TextSelected)

	return a
}

func (a *AlbumView) SetAlbum(album *models.Album, songs []*models.Song) {
	a.list.Clear()
	a.songs = make([]*albumSong, len(songs))
	items := make([]twidgets.ListItem, len(songs))

	album.SongCount = len(a.songs)
	a.album = album

	text := album.Name
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
	}

	a.list.AddItems(items...)
}

func (a *AlbumView) SetArtist(artist *models.Artist) {
	a.artist = artist
}

func (a *AlbumView) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		key := event.Key()
		if a.listFocused {
			index := a.list.GetSelectedIndex()
			if index == 0 && (key == tcell.KeyUp || key == tcell.KeyCtrlK) {
				a.listFocused = false
				a.prevBtn.Focus(func(p cview.Primitive) {})
				a.list.Blur()
			} else if key == tcell.KeyEnter && event.Modifiers() == 0 {
				a.playSong(index)
			} else {
				a.list.InputHandler()(event, setFocus)
			}
		} else {
			if key == tcell.KeyDown || key == tcell.KeyCtrlJ {
				a.listFocused = true
				a.list.Focus(func(p cview.Primitive) {})
			} else {
			}
		}
	}
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

func (a *AlbumView) listHandler(key *tcell.EventKey) *tcell.EventKey {
	if key.Key() == tcell.KeyEnter && key.Modifiers() == tcell.ModNone {
		index := a.list.GetSelectedIndex()
		a.playSong(index)
		return nil
	}
	return key
}

func (a *AlbumView) showSimilar() {
	if a.similarFunc != nil {
		a.similarFunc(a.album)
	}
}
