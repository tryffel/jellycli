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
	"github.com/sirupsen/logrus"
	"sync"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/controller"
	"tryffel.net/pkg/jellycli/models"
)

type panelSplit int

const (
	panelL panelSplit = iota
	panelR
)

func (p panelSplit) Other() panelSplit {
	if p == panelL {
		return panelR
	} else if p == panelR {
		return panelL
	} else {
		return -1
	}
}

type browserTransition int

const (
	transitionSelectArtist browserTransition = iota
	transitionShowAlbums
	transitionSelectAlbum
	transitionShowSongs
	transitionReset
)

type browserState int

const (
	stateArtists browserState = iota
	stateArtistAlbums
	stateAlbumSongs
)

type browserAction int

const (
	//Enter, for artist show albums, for albums show songs
	actionEnter browserAction = iota
	//Back, for songs go to albums, for albums go to artists
	actionBack
)

// Browser is a listR-like viewer user can navigate content with
type Browser struct {
	// Widgets
	grid  *tview.Grid
	listR *List
	listL *List

	controller controller.MediaController

	// State
	rContent      models.ItemType
	lContent      models.ItemType
	panelAwaiting panelSplit

	dataL   []models.Item
	dataR   []models.Item
	element models.ItemType

	hasModal   bool
	gridAxis   []int
	gridSize   int
	customGrid bool
	focused    panelSplit
	modal      Modal

	transition browserTransition
	state      browserState

	lock sync.RWMutex
}

func (b *Browser) Draw(screen tcell.Screen) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	b.grid.Draw(screen)
}

func (b *Browser) GetRect() (int, int, int, int) {
	return b.grid.GetRect()
}

func (b *Browser) SetRect(x, y, width, height int) {
	b.grid.SetRect(x, y, width, height)
}

func (b *Browser) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if b.hasModal {
			b.modal.View().InputHandler()(event, setFocus)
			return
		}

		if event.Key() == tcell.KeyTAB {
			if b.focused == panelL {
				b.focused = panelR
				b.listR.Focus(nil)
				b.listL.Blur()
			} else {
				b.focused = panelL
				b.listL.Focus(nil)
				b.listR.Blur()
			}
			return
		}

		//if event.Key() == tcell.KeyEnter {
		if b.focused == panelL {
			b.listL.InputHandler()(event, setFocus)
		} else {
			b.listR.InputHandler()(event, setFocus)
		}
	}
}

func (b *Browser) Focus(delegate func(p tview.Primitive)) {
	if b.focused == panelL {
		b.listL.Focus(delegate)
		b.listR.Blur()
	} else {
		b.listR.Focus(delegate)
		b.listL.Blur()
	}
	//b.grid.Focus(delegate)
}

func (b *Browser) Blur() {
	b.listL.Blur()
	b.listR.Blur()
	b.grid.Blur()
}

func (b *Browser) GetFocusable() tview.Focusable {
	return b.grid.GetFocusable()
}

func (b *Browser) setData(data []models.Item) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.panelAwaiting == panelR {
		b.listR.SetData(data)
		b.dataR = data
	} else if b.panelAwaiting == panelL {
		b.listL.SetData(data)
		b.dataL = data
	} else {
		var itemType string
		if data != nil {
			if len(data) > 0 {
				itemType = string(data[0].GetType())
			} else {
				itemType = "empty slice"
			}
		} else {
			itemType = "nil"
		}
		logrus.Error("Browser received items it wasn't waiting for, type: ", itemType)
		return
	}
	b.panelAwaiting = -1
}

func NewBrowser(controller controller.MediaController) *Browser {
	b := &Browser{
		listR:      nil,
		listL:      nil,
		grid:       tview.NewGrid(),
		controller: controller,
	}

	b.listR = NewList(b.wrapEnter(panelR))
	b.listL = NewList(b.wrapEnter(panelL))

	b.listL.itemType = models.TypeArtist
	b.listR.itemType = models.TypeAlbum

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

	b.transition = transitionReset
	b.state = stateArtists
	b.panelAwaiting = panelL

	//b.controller.SetItemsCallback(b.setData)
	return b
}

//AddModal adds modal to center of browser
func (b *Browser) AddModal(modal Modal, height, width uint, lockSize bool) {
	if b.hasModal {
		return
	}
	if !lockSize {
		b.customGrid = false
		b.grid.AddItem(modal.View(), 2, 2, 2, 2, 8, 30, true)
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
		b.grid.AddItem(modal.View(), 2, 2, 2, 2, int(height), int(width), true)
	}
	b.hasModal = true
	b.modal = modal
}

//RemoveModal removes modal
func (b *Browser) RemoveModal(view tview.Primitive) {
	if b.hasModal {
		b.grid.RemoveItem(view)
		b.hasModal = false
		b.modal = nil
		if b.customGrid {
			b.grid.SetRows(b.gridAxis...)
			b.grid.SetColumns(b.gridAxis...)
			b.customGrid = false
		}
	}
}

func (b *Browser) wrapEnter(panel panelSplit) func(index int) {
	return func(index int) {
		b.enter(panel, index)
	}
}

func (b *Browser) enter(panel panelSplit, index int) {
	index, item := b.getSelectedItem(panel)
	logrus.Debug("Selected: ", item.GetName())

	// Play song
	if item.GetType() == models.TypeSong {
		song := item.(*models.Song)
		b.controller.AddSongs([]*models.Song{song})

	} else {
		// Update browser view
		b.makeTransition(panel, tcell.KeyEnter)
	}
}

func (b *Browser) makeTransition(panel panelSplit, key tcell.Key) {
	/* State transitions
	Possible item types:
	Left panel: album, artist
	Right panel: playlist, album, song, queue, history

	On left panel & enter open right panel with corresponding content
	On right panel if album:
		move album to left
		show songs on right
	else:
		play

	*/

	action := browserAction(-1)
	if (key == tcell.KeyEnter) || (key == tcell.KeyRight) {
		action = actionEnter
	} else if key == tcell.KeyLeft {
		action = actionBack
	} else {
		return
	}

	if action == actionEnter {
		err := b.transitionEnter(panel)
		if err != nil {
			logrus.Error("Error transitioning enter on browser: ", err)
		}

		//index, item := b.getSelectedItem(panel)

	}

	/*
		if panel == panelL {
			if key == tcell.KeyEnter {
				index := b.listL.list.GetCurrentItem()
				switch b.lContent {
				case models.TypeAlbum:
					b.panelAwaiting = panelR
					b.controller.GetItem(b.dataL[index].GetParent(), b.dataL[index].GetType())
				}
			}

		} else if panel == panelR {

		}

	*/

}

func (b *Browser) getSelectedItem(split panelSplit) (int, models.Item) {
	var index int
	var item models.Item
	if split == panelL {
		index = b.listL.list.GetCurrentItem()
		item = b.dataL[index]
	} else if split == panelR {
		index = b.listR.list.GetCurrentItem()
		item = b.dataR[index]
	}
	return index, item
}
