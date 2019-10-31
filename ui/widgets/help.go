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
	"strings"
	"tryffel.net/pkg/jellycli/config"
)

type Help struct {
	grid    *tview.Grid
	header  *tview.TextView
	text    *tview.TextView
	logo    *tview.TextView
	visible bool
	closeCb func()
}

func (h *Help) SetDoneFunc(doneFunc func()) {
	h.closeCb = doneFunc
}

func (h *Help) View() tview.Primitive {
	return h
}

func (h *Help) SetVisible(visible bool) {
	h.visible = visible

}

func (h *Help) Draw(screen tcell.Screen) {
	h.grid.Draw(screen)
}

func (h *Help) GetRect() (int, int, int, int) {
	return h.grid.GetRect()
}

func (h *Help) SetRect(x, y, width, height int) {
	h.grid.SetRect(x, y, width, height)
}

func (h *Help) InputHandler() func(event *tcell.EventKey, setfocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setfocus func(p tview.Primitive)) {
		if h.visible && event.Key() == tcell.KeyEscape {
			if h.closeCb != nil {
				h.closeCb()
			}
		}
	}
}

func (h *Help) Focus(delegate func(p tview.Primitive)) {
	h.grid.SetBorderColor(config.ColorBorderFocus)
	h.grid.Focus(delegate)
}

func (h *Help) Blur() {
	h.grid.SetBorderColor(config.ColorBorder)
	h.grid.Blur()
}

func (h *Help) GetFocusable() tview.Focusable {
	return h.grid.GetFocusable()
}

func NewHelp(doneCb func()) *Help {
	h := &Help{text: tview.NewTextView()}
	h.closeCb = doneCb

	h.grid = tview.NewGrid()
	h.grid.SetRows(7, 0)
	h.grid.SetColumns(-1)
	h.grid.SetBackgroundColor(config.ColorBackground)
	h.grid.SetBorder(true)
	h.grid.SetTitle("Help")
	h.grid.SetBorderColor(config.ColorBorder)
	h.grid.SetTitleColor(config.ColorPrimary)
	h.grid.SetBorderPadding(0, 1, 2, 2)
	h.grid.SetMinSize(6, 6)
	//h.grid.SetGap(3,3)
	config.DebugGridBorders(h.grid)

	h.logo = tview.NewTextView()
	h.logo.SetBorder(false)
	h.logo.SetBackgroundColor(config.ColorBackground)
	h.logo.SetTextColor(config.ColorPrimary)
	h.logo.SetTextAlign(tview.AlignCenter)
	h.logo.SetWrap(false)
	h.logo.SetWordWrap(false)

	h.text.SetBorder(false)
	h.text.SetBackgroundColor(config.ColorBackground)
	h.text.SetTextColor(config.ColorPrimary)
	h.text.SetWordWrap(true)
	h.text.SetTextAlign(tview.AlignLeft)

	_, _ = h.logo.Write([]byte(logo()))
	_, _ = h.logo.Write([]byte(fmt.Sprintf("\n v%s", config.Version)))
	_, _ = h.text.Write([]byte(helpText()))

	h.grid.AddItem(h.logo, 0, 0, 1, 3, 6, 40, false)
	h.grid.AddItem(h.text, 1, 0, 1, 3, 6, 30, false)
	return h
}

func logo() string {
	text := `
   __         _  _               _  _ 
   \ \   ___ | || | _   _   ___ | |(_)
    \ \ / _ \| || || | | | / __|| || |
 /\_/ /|  __/| || || |_| || (__ | || |
 \___/  \___||_||_| \__, | \___||_||_|
                    |___/`
	return strings.TrimLeft(text, "\n")
}

func helpText() string {
	return `Help page for Jellycli. 
Press Escape to return

`
}
