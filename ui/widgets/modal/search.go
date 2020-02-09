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

package modal

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"tryffel.net/go/jellycli/config"
)

type Search struct {
	grid      *tview.Grid
	okBtn     *tview.Button
	cancelBtn *tview.Button
	input     *tview.InputField
	visible   bool

	searchCb  func(string, bool)
	closeFunc func()
	selected  int
}

func (s *Search) SetDoneFunc(doneFunc func()) {
	s.closeFunc = doneFunc
}

func (s *Search) View() tview.Primitive {
	return s
}

func (s *Search) SetVisible(visible bool) {
	s.visible = visible
}

func (s *Search) Draw(screen tcell.Screen) {
	s.grid.Draw(screen)
}

func (s *Search) GetRect() (int, int, int, int) {
	return s.grid.GetRect()
}

func (s *Search) SetRect(x, y, width, height int) {
	s.grid.SetRect(x, y, width, height)
}

func (s *Search) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return s.grid.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		var search = 0
		var ok = 1
		var cancel = 2
		key := event.Key()
		if !s.visible {
			return
		}
		if key == config.KeyBinds.Moving.Down || key == config.KeyBinds.Moving.DownAlt {
			if s.selected == search {
				s.selected = ok
				//setFocus = s.okBtn
			} else if s.selected == ok {
				s.selected = cancel

			}

		} else if key == config.KeyBinds.Moving.Up || key == config.KeyBinds.Moving.UpAlt {
			if s.selected == cancel {
				s.selected = ok
			} else if s.selected == ok {
				s.selected = search
			}
		} else if key == tcell.KeyEnter {
			s.searchCb(s.input.GetText(), true)
		} else if key == tcell.KeyEsc {
			s.searchCb(s.input.GetText(), false)
		}
	})
}

func (s *Search) Focus(delegate func(p tview.Primitive)) {
	s.grid.SetBorderColor(config.ColorBorderFocus)
	s.grid.Focus(delegate)
	switch s.selected {
	case 0:
		s.input.Focus(delegate)
	case 1:
		s.okBtn.Focus(delegate)
	case 2:
		s.cancelBtn.Focus(delegate)
	default:
		s.grid.Focus(delegate)
	}
}

func (s *Search) Blur() {
	s.grid.SetBorderColor(config.ColorBorder)
	s.grid.Blur()
}

func (s *Search) GetFocusable() tview.Focusable {
	return s.grid.GetFocusable()
}

func NewSearch(searchCb func(string, bool)) *Search {
	s := &Search{
		grid:      tview.NewGrid(),
		okBtn:     tview.NewButton("Search"),
		cancelBtn: tview.NewButton("Cancel"),
		input:     tview.NewInputField(),
	}

	s.selected = 0
	s.okBtn.SetLabel("Search")
	s.cancelBtn.SetLabel("Cancel")

	s.grid.SetRows(-1, 1, 1, 1, -1)
	s.grid.SetColumns(-2, 8, -1, 8, -2)
	s.grid.SetBackgroundColor(config.ColorBackground)
	s.grid.SetBorder(true)
	s.grid.SetTitle("Search")
	s.grid.SetBorderColor(config.ColorBorder)
	s.grid.SetTitleColor(config.ColorPrimary)
	s.grid.SetBorderPadding(1, 1, 2, 2)
	s.grid.SetMinSize(1, 4)
	config.DebugGridBorders(s.grid)

	s.input.SetLabel("Search: ")
	s.input.SetPlaceholder("E.g. 'africa'")
	s.input.SetFieldWidth(50)
	s.input.SetAcceptanceFunc(s.acceptFunc)
	s.input.SetTitle("Search")
	s.input.SetTitleColor(config.ColorPrimary)
	s.input.SetLabelColor(config.ColorSecondary)
	s.input.SetLabelWidth(10)

	s.input.SetDoneFunc(s.doneFunc)

	s.okBtn.SetBackgroundColor(config.ColorControls)
	s.okBtn.SetLabelColor(config.ColorBackground)
	s.cancelBtn.SetBackgroundColor(config.ColorControls)
	s.cancelBtn.SetLabelColor(config.ColorBackground)

	s.searchCb = searchCb

	s.grid.AddItem(s.input, 1, 0, 1, 5, 1, 30, true)
	s.grid.AddItem(s.okBtn, 3, 1, 1, 1, 1, 4, false)
	s.grid.AddItem(s.cancelBtn, 3, 3, 1, 1, 1, 4, false)
	return s
}

func (s *Search) acceptFunc(tect string, ch rune) bool {
	return true
}

func (s *Search) doneFunc(key tcell.Key) {
	s.Blur()
	if key == tcell.KeyEnter {
		s.search()
	} else if key == tcell.KeyEsc {
		s.cancel()
	}
	s.closeFunc()
}

func (s *Search) search() {
	if s.searchCb != nil {
		text := s.input.GetText()
		s.searchCb(text, true)
	}
}

func (s *Search) cancel() {
	if s.searchCb != nil {
		text := s.input.GetText()
		s.searchCb(text, false)
	}
}
