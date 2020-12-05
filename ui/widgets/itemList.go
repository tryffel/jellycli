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
	"gitlab.com/tslocum/cview"
	"strings"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/twidgets"
)

// characters to strip from search texts and search input
const stripCharacters = "-+_.:,;&#%!'"

// itemList shows Banner (title, buttons) and list below header.
// It also features filtering list items.
//
// To enable filtering, set itemList.reduceEnabled = true.
// Reduce filters texts in itemList.itemsText to find correct items.
// Reduce also calls setReducerVisibile when reduce input should be shown/hidden
// and it is users responsibility to add/remove it from grid. After this, itemList hands
// focus to reduceInput.
// When user selects item from filtered (reduced) list, listSelectFunc is called
// with items original index, not reduced index.
// If there are external changes to item list (such as paging, refresh etc),
// call itemList.resetReduce to reset reducer state.
type itemList struct {
	*twidgets.Banner
	*previous
	list        *twidgets.ScrollList
	listFocused bool

	description *cview.TextView
	prevBtn     *button

	prevFunc func()

	reduceInput   *cview.InputField
	reduceVisible bool
	reduceEnabled bool

	setReducerVisible func(bool)
	itemsTexts        []string
	items             []twidgets.ListItem
	reduceIndices     []int

	listSelectFunc func(index int)
}

func newItemList(listSelectfunc func(index int)) *itemList {
	itemList := &itemList{
		Banner:   twidgets.NewBanner(),
		previous: &previous{},

		description: cview.NewTextView(),
		prevBtn:     newButton("Back"),
		prevFunc:    nil,
	}

	itemList.list = twidgets.NewScrollList(itemList.selectitem)
	itemList.listSelectFunc = listSelectfunc

	rInput := cview.NewInputField()
	itemList.reduceInput = rInput

	rInput.SetBorder(true)
	rInput.SetBorderColor(config.Color.Border)
	rInput.SetChangedFunc(itemList.reduce)
	// leave space for printing num of results
	rInput.SetLabel("Filter ")
	rInput.SetLabelWidth(13)
	rInput.SetFieldBackgroundColor(config.Color.BackgroundSelected)
	rInput.SetFieldTextColor(config.Color.TextSelected)

	itemList.SetBorder(true)
	itemList.SetBackgroundColor(config.Color.Background)
	itemList.Grid.SetBackgroundColor(config.Color.Background)
	itemList.list.SetBackgroundColor(config.Color.Background)

	itemList.description.SetDynamicColors(true)

	itemList.list.SetBorder(true)
	itemList.list.Grid.SetColumns(1, -1)

	itemList.SetBorder(true)
	itemList.prevBtn.SetSelectedFunc(itemList.goBack)

	itemList.list.PreInputHandler = itemList.InputHandler()
	return itemList
}

// init context menu list. Context menu list has to contain at least one item
// before calling this.
func (i *itemList) initContextMenuList() {
	i.list.ContextMenuList().SetBorder(true)
	i.list.ContextMenuList().SetBackgroundColor(config.Color.Background)
	i.list.ContextMenuList().SetBorderColor(config.Color.BorderFocus)
	i.list.ContextMenuList().SetSelectedBackgroundColor(config.Color.BackgroundSelected)
	i.list.ContextMenuList().SetMainTextColor(config.Color.Text)
	i.list.ContextMenuList().SetSelectedTextColor(config.Color.TextSelected)
}

func (i *itemList) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		r := event.Rune()
		if r == ' ' {
			if i.reduceEnabled && config.AppConfig.Gui.EnableResultsFiltering {
				if i.setReducerVisible != nil {
					i.setReducerVisible(true)
					i.reduceVisible = true
				}

				i.reduceInput.SetDoneFunc(i.reducerDone(setFocus))
				i.reduceInput.SetFieldBackgroundColor(config.Color.BackgroundSelected)
				setFocus(i.reduceInput)
			}
		}
	}
}

func (i *itemList) reducerDone(setFocus func(p cview.Primitive)) func(key tcell.Key) {
	return func(key tcell.Key) {
		// hand focus over to list, either with reduced results or reduce reset first
		if i.reduceVisible {
			switch key {
			case tcell.KeyEsc:
				if i.setReducerVisible != nil {
					i.setReducerVisible(false)
					i.resetReduce()
					i.reduceVisible = false
				}
				setFocus(i.list)
			case tcell.KeyEnter:
				setFocus(i.list)
				i.reduceInput.SetFieldBackgroundColor(config.Color.Background)
			}
		}
	}
}

func (i *itemList) reduce(input string) {
	// Find items that match given text. Simply iterating over all results
	// seems to be fast enough for this use case, where list if (hopefully)
	// < 1000 items.

	lowerCase := strings.ToLower(input)
	rawTokens := strings.Split(lowerCase, " ")
	tokens := make([]string, 0, 1)
	for _, v := range rawTokens {
		stripped := strings.Trim(v, stripCharacters)
		if len(v) > 0 {
			tokens = append(tokens, stripped)
		}
	}

	if len(i.items) > 0 {
		selected := i.list.GetSelectedIndex()
		i.items[selected].SetSelected(twidgets.Deselected)
	}

	indices := make([]int, 0, 10)
	for index, v := range i.itemsTexts {
		tokenFound := true
		for _, token := range tokens {
			if !strings.Contains(v, token) {
				tokenFound = false
				break
			}
		}
		if tokenFound {
			indices = append(indices, index)
		}
	}
	items := make([]twidgets.ListItem, len(indices))
	for index, v := range indices {
		items[index] = i.items[v]
	}
	i.list.Clear()
	i.list.AddItems(items...)
	i.reduceIndices = indices
	i.reduceInput.SetLabel(fmt.Sprintf("Filter (%d)", len(items)))
}

func (i *itemList) searchItemsSet() {
	for iText, v := range i.itemsTexts {
		lower := strings.ToLower(v)
		stripped := strings.Trim(lower, stripCharacters)
		i.itemsTexts[iText] = stripped
	}
}

func (i *itemList) drawReducedResultsCount(screen tcell.Screen) {
	x, y, w, h := i.reduceInput.GetInnerRect()
	if w > 40 && h >= 1 && i.reduceInput.GetText() != "" {
		text := fmt.Sprintf("%d results", len(i.items))
		xStart := x + w - len(text)*2
		cview.Print(screen, text, xStart, y, 20, cview.AlignRight, config.Color.TextSecondary)
	}
}

func (i *itemList) resetReduce() {
	if i.reduceVisible {
		if len(i.items) > 0 {
			selected := i.list.GetSelectedIndex()
			i.items[selected].SetSelected(twidgets.Deselected)
		}
		i.reduceInput.SetText("")
		i.reduceInput.SetLabel("Filter")
		i.list.Clear()
		i.list.AddItems(i.items...)
		i.setReducerVisible(false)
	}
}

func (i *itemList) selectitem(index int) {
	// in case of reduce, return original items index, not reduced index.
	if i.reduceVisible {
		index = i.reduceIndices[index]
	}
	i.listSelectFunc(index)
}

func (i *itemList) getSelectedIndex() int {
	var index int
	if i.reduceVisible {
		index = i.reduceIndices[index]
	} else {
		index = i.list.GetSelectedIndex()
	}
	return index
}
