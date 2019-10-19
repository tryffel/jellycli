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

// Browser is a listR-like viewer user can navigate content with
type Browser struct {
	grid  *tview.Grid
	listR *List
	listL *List

	data    []models.Artist
	element models.ListElement

	hasModal   bool
	gridAxis   []int
	gridSize   int
	customGrid bool
}

func (b *Browser) Draw(screen tcell.Screen) {
	b.grid.Draw(screen)
}

func (b *Browser) GetRect() (int, int, int, int) {
	return b.grid.GetRect()
}

func (b *Browser) SetRect(x, y, width, height int) {
	b.grid.SetRect(x, y, width, height)
}

func (b *Browser) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.grid.InputHandler()
}

func (b *Browser) Focus(delegate func(p tview.Primitive)) {
	b.grid.Focus(delegate)
}

func (b *Browser) Blur() {
	b.grid.Blur()
}

func (b *Browser) GetFocusable() tview.Focusable {
	return b.grid.GetFocusable()
}

func (b *Browser) setData(data *[]models.Item, element models.ListElement) {
	b.listL.SetData(element, *data)
	b.listR.SetData((*data)[0].GetChildren()[0].GetType(), (*data)[0].GetChildren())
}

func NewBrowser() *Browser {
	b := &Browser{data: nil, listR: NewList(), listL: NewList(), grid: tview.NewGrid()}

	config.DebugGridBorders(b.grid)
	// Split grid to 6x6, normally use 3x6 panes for both lists
	// When showing modal on top, put it to central 4 cells

	b.gridAxis = []int{2, -1, -2, -2, -1, 2}
	b.gridSize = 6
	b.grid.SetRows(b.gridAxis...)
	b.grid.SetColumns(b.gridAxis...)

	b.grid.SetMinSize(2, 2)
	b.grid.SetBackgroundColor(config.ColorBackground)

	b.grid.AddItem(b.listL, 0, 0, b.gridSize, b.gridSize/2, 4, 10, false)
	b.grid.AddItem(b.listR, 0, b.gridSize/2, b.gridSize, b.gridSize/2, 4, 10, false)
	return b
}

//AddModal adds modal to center of browser
func (b *Browser) AddModal(view tview.Primitive, height, width uint, lockSize bool) {
	if b.hasModal {
		return
	}
	if !lockSize {
		b.customGrid = false
		b.grid.AddItem(view, 2, 2, 2, 2, 8, 30, true)
	} else {
		b.customGrid = true
		x := make([]int, len(b.gridAxis))
		y := make([]int, len(b.gridAxis))
		copy(x, b.gridAxis)
		copy(y, b.gridAxis)
		x[2] = int(width / 2)
		x[3] = x[2]
		y[2] = int(height / 2)
		y[3] = y[2]
		b.grid.SetRows(y...)
		b.grid.SetColumns(x...)
		b.grid.AddItem(view, 2, 2, 2, 2, int(height), int(width), true)
	}
	b.hasModal = true
}

//RemoveModal removes modal
func (b *Browser) RemoveModal(view tview.Primitive) {
	if b.hasModal {
		b.grid.RemoveItem(view)
		b.hasModal = false

		if b.customGrid {
			b.grid.SetRows(b.gridAxis...)
			b.grid.SetColumns(b.gridAxis...)
			b.customGrid = false
		}
	}
}
