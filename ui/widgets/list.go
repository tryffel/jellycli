/*
 * Copyright 2019 Tero Vierimaa
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
	"tryffel.net/pkg/jellycli/models"
)

type List struct {
	list     *tview.List
	itemType models.ItemType
	items    []models.Item
	enterCb  func(index int)
}

func NewList(enterCb func(index int)) *List {
	l := &List{
		list: tview.NewList(),
	}
	l.list.SetBorder(true)
	l.list.SetBorderColor(config.ColorBorder)
	l.list.SetTitleColor(config.ColorBorder)
	l.list.SetTitleAlign(tview.AlignLeft)
	l.list.ShowSecondaryText(false)
	l.list.SetShortcutColor(tcell.ColorDefault)
	l.list.SetBackgroundColor(config.ColorBackground)
	l.list.SetSelectedTextColor(config.ColorSecondary)
	l.list.SetMainTextColor(config.ColorPrimary)
	l.list.SetHighlightFullLine(true)
	l.enterCb = enterCb
	return l
}

func (l *List) Draw(screen tcell.Screen) {
	l.list.Draw(screen)
}

func (l *List) GetRect() (int, int, int, int) {
	return l.list.GetRect()
}

func (l *List) SetRect(x, y, width, height int) {
	l.list.SetRect(x, y, width, height)
}

func (l *List) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEnter && l.enterCb != nil {
			index := l.list.GetCurrentItem()
			l.enterCb(index)
		}
		l.list.InputHandler()(event, setFocus)
	}
}

func (l *List) Focus(delegate func(p tview.Primitive)) {
	l.list.SetBorderColor(config.ColorBorderFocus)
	l.list.Focus(delegate)
}

func (l *List) Blur() {
	l.list.SetBorderColor(config.ColorBorder)
	l.list.Blur()
}

func (l *List) GetFocusable() tview.Focusable {
	return l.list.GetFocusable()
}

func (l *List) SetData(items []models.Item) {
	if len(items) == 0 {
		return
	}

	l.itemType = items[0].GetType()
	l.list.Clear()
	l.list.SetTitle(fmt.Sprintf("%ss", l.itemType))
	switch l.itemType {
	case models.TypeArtist:
		for i, v := range items {
			text := ""
			artist, ok := v.(*models.Artist)
			if ok {
				text = fmt.Sprintf("%s - %s", artist.Name, SecToString(artist.TotalDuration))
			} else {
				text = v.GetName()
			}
			l.list.AddItem(text, "", 0, l.namedCb(i, string(v.GetId())))
		}
	case models.TypeAlbum:
		for i, v := range items {
			text := ""
			album, ok := v.(*models.Album)
			if ok {
				text = fmt.Sprintf("%s, %d - %s", album.Name, album.Year, SecToString(album.Duration))
			} else {
				text = v.GetName()
			}
			l.list.AddItem(text, "", 0, l.namedCb(i, string(v.GetId())))
		}
	case models.TypeSong:
		for i, v := range items {
			text := ""
			song, ok := v.(*models.Song)
			if ok {
				text = fmt.Sprintf("%s - %s", song.Name, SecToString(song.Duration))
			} else {
				text = v.GetName()
			}
			l.list.AddItem(text, "", 0, l.namedCb(i, string(v.GetId())))
		}

	default:
		for i, v := range items {
			l.list.AddItem(v.GetName(), "", 0, l.namedCb(i, string(v.GetId())))
		}
	}
}

func (l *List) namedCb(i int, id string) func() {
	return func() {
		l.selectCb(i, id)
	}
}

func (l *List) selectCb(i int, id string) {

}

func (l *List) Clear() {
	l.list.Clear()
}
