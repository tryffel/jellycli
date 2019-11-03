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
	"unicode/utf8"
)

type List struct {
	table    *tview.Table
	itemType models.ItemType
	items    []models.Item
	enterCb  func(index int)
}

func NewList(enterCb func(index int)) *List {
	l := &List{
		table: tview.NewTable(),
	}
	l.enterCb = enterCb

	l.table.SetBorder(true)
	l.table.SetBorders(false)
	l.table.SetBorderColor(config.ColorBorder)
	l.table.SetTitleColor(config.ColorBorder)
	l.table.SetTitleAlign(tview.AlignLeft)
	l.table.SetBackgroundColor(config.ColorBackground)
	l.table.SetSelectedStyle(config.ColorPrimary, config.ColorBorder, 0)
	l.table.SetSelectable(true, false)
	l.table.SetFixed(1, 10)

	return l
}

func (l *List) Draw(screen tcell.Screen) {
	l.table.Draw(screen)
}

func (l *List) GetRect() (int, int, int, int) {
	return l.table.GetRect()
}

func (l *List) SetRect(x, y, width, height int) {
	l.table.SetRect(x, y, width, height)
}

func (l *List) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEnter && l.enterCb != nil {
			index, _ := l.table.GetSelection()
			if index > 0 {
				l.enterCb(index)
			}
		}
		l.table.InputHandler()(event, setFocus)
	}
}

func (l *List) Focus(delegate func(p tview.Primitive)) {
	l.table.SetBorderColor(config.ColorBorderFocus)
	l.table.Focus(delegate)
}

func (l *List) Blur() {
	l.table.SetBorderColor(config.ColorBorder)
	l.table.Blur()
}

func (l *List) GetFocusable() tview.Focusable {
	return l.table.GetFocusable()
}

func (l *List) SetData(items []models.Item) {
	if len(items) == 0 {
		return
	}

	l.itemType = items[0].GetType()
	l.table.Clear()
	l.table.SetCell(0, 0, tableHeaderCell("#"))
	switch l.itemType {
	case models.TypeArtist:
		l.table.SetCell(0, 1, tableHeaderCell("Artist"))
		l.table.SetCell(0, 2, tableHeaderCell("Duration"))
		for i, v := range items {
			l.table.SetCell(i+1, 0, tableCell(fmt.Sprintf("%d.", i+1)))
			artist, ok := v.(*models.Artist)
			if ok {
				l.table.SetCell(i+1, 1, setCellWidth(tableCell(artist.Name)))
				l.table.SetCell(i+1, 2, tableCell(SecToStringApproximate(artist.TotalDuration)))
			} else {
				l.table.SetCell(i+1, 1, setCellWidth(tableCell(v.GetName()+"(unknown type)")))
			}
		}
	case models.TypeAlbum:
		l.table.SetCell(0, 1, tableHeaderCell("Album"))
		l.table.SetCell(0, 2, tableHeaderCell("Year"))
		l.table.SetCell(0, 3, tableHeaderCell("Duration"))
		for i, v := range items {
			l.table.SetCell(i+1, 0, tableCell(fmt.Sprintf("%d.", i+1)))
			album, ok := v.(*models.Album)
			if ok {
				l.table.SetCell(i+1, 1, setCellWidth(tableCell(album.Name)))
				l.table.SetCell(i+1, 2, tableCell(fmt.Sprint(album.Year)))
				l.table.SetCell(i+1, 3, tableCell(SecToStringApproximate(album.Duration)))
			} else {
				l.table.SetCell(i+1, 1, setCellWidth(tableCell(v.GetName()+"(unknown type)")))
			}
		}
	case models.TypeSong:
		l.table.SetCell(0, 1, tableHeaderCell("Song"))
		l.table.SetCell(0, 2, tableHeaderCell("Duration"))
		for i, v := range items {
			l.table.SetCell(i+1, 0, tableCell(fmt.Sprintf("%d.", i+1)))
			song, ok := v.(*models.Song)
			if ok {
				l.table.SetCell(i+1, 1, tableCell(song.Name))
				l.table.SetCell(i+1, 2, tableCell(SecToString(song.Duration)))
			} else {
				l.table.SetCell(i+1, 1, setCellWidth(tableCell(v.GetName()+"(unknown type)")))
			}
		}

	default:
		for i, v := range items {
			l.table.SetCell(i, 1, tableCell(v.GetName()))
		}
	}

	l.table.Select(1, 0)
}

//SelectedIndex returns currently selected index
func (l *List) SelectedIndex() int {
	index, _ := l.table.GetSelection()
	return index - 1

}

func (l *List) namedCb(i int, id string) func() {
	return func() {
		l.selectCb(i, id)
	}
}

func (l *List) selectCb(i int, id string) {

}

func (l *List) Clear() {
	l.table.Clear()
}

func tableCell(text string) *tview.TableCell {
	c := tview.NewTableCell(text)
	c.SetTextColor(config.ColorPrimary)
	c.SetAlign(tview.AlignLeft)
	return c
}

func tableHeaderCell(text string) *tview.TableCell {
	c := tview.NewTableCell(text)
	c.SetTextColor(config.ColorSecondary)
	c.SetAlign(tview.AlignLeft)
	c.SetSelectable(false)
	return c
}

func setCellWidth(cell *tview.TableCell) *tview.TableCell {
	want := 25
	cell.SetMaxWidth(want)
	length := utf8.RuneCountInString(cell.Text)
	need := want - length
	ten := "          "
	five := "     "

	if need < 5 {
		return cell
	} else if need < 10 {
		cell.Text += five
	} else if need < 15 {
		cell.Text += ten
	} else if need < 20 {
		cell.Text += ten + five
	}
	return cell
}
