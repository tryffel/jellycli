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
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/models"
)

func testData() *[]models.Item {

	data := &[]models.Artist{}

	items := make([]models.Item, len(*data))
	for i, v := range *data {
		items[i] = models.Item(&v)
	}
	return &items
}

// Browser is a list-like viewer user can navigate content with
type Browser struct {
	list *List

	data    []models.Artist
	element models.ListElement
}

func (b *Browser) Draw(screen tcell.Screen) {
	b.list.list.SetBackgroundColor(tcell.ColorDefault)
	b.list.Draw(screen)
}

func (b *Browser) GetRect() (int, int, int, int) {
	return b.list.GetRect()
}

func (b *Browser) SetRect(x, y, width, height int) {
	b.list.SetRect(x, y, width, height)
}

func (b *Browser) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.InputHandler()
}

func (b *Browser) Focus(delegate func(p tview.Primitive)) {
	b.list.Focus(delegate)
}

func (b *Browser) Blur() {
	b.list.Blur()
}

func (b *Browser) GetFocusable() tview.Focusable {
	return b.list.GetFocusable()
}

func (b *Browser) setData(data *[]models.Item, element models.ListElement) {

	b.list.SetData(element, *data)
}

func NewBrowser() *Browser {
	b := &Browser{data: nil, list: NewList()}

	b.list.list.SetBorder(true)
	b.list.list.SetBorderColor(config.ColorBorder)
	b.list.list.SetTitleColor(config.ColorBorder)
	b.list.list.SetTitleAlign(tview.AlignLeft)
	b.list.list.ShowSecondaryText(false)
	b.list.list.SetShortcutColor(tcell.ColorDefault)
	return b

}
