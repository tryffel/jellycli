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
)

type NavBar struct {
	grid *tview.Grid
	btns []*tview.Button
}

func (n *NavBar) Draw(screen tcell.Screen) {
	n.grid.Draw(screen)
}

func (n *NavBar) GetRect() (int, int, int, int) {
	return n.grid.GetRect()
}

func (n *NavBar) SetRect(x, y, width, height int) {
	n.grid.SetRect(x, y, width, height)
}

func (n *NavBar) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return n.grid.InputHandler()
}

func (n *NavBar) Focus(delegate func(p tview.Primitive)) {
	n.grid.Focus(delegate)
}

func (n *NavBar) Blur() {
	n.grid.Blur()
}

func (n *NavBar) GetFocusable() tview.Focusable {
	return n.grid.GetFocusable()
}

func NewNavBar() *NavBar {
	nb := &NavBar{
		grid: tview.NewGrid(),
		btns: nil,
	}

	nb.grid.SetBorder(false)
	nb.grid.SetBorderColor(config.ColorNavBar)
	nb.grid.SetBackgroundColor(config.ColorNavBar)
	nb.grid.SetRows(-1)

	buttons := []string{
		"Help",
		"Search",
		"Queue",
		"History",
		"Settings",
		"Quit",
	}

	// Use grid of |<space><button>space><button><space>...|
	widths := make([]int, len(buttons)*2-1)
	spaceWidth := -1
	for i, _ := range buttons {
		widths[i*2] = -2
		if i > 0 {
			widths[i*2-1] = spaceWidth
		}
	}

	nb.grid.SetColumns(widths...)
	nb.btns = make([]*tview.Button, len(buttons))
	for i, name := range buttons {
		nb.btns[i] = tview.NewButton(name)
		nb.btns[i].SetSelectedFunc(nb.namedCb(name))
		nb.btns[i].SetBackgroundColor(config.ColorNavBar)
		nb.btns[i].SetLabelColor(config.ColorLightext)
		nb.grid.AddItem(nb.btns[i], 0, i*2, 1, 1, 1, 4, false)
	}
	return nb
}

func (n *NavBar) namedCb(name string) func() {
	return func() {
		n.buttonCb(name)
	}
}

func (n *NavBar) buttonCb(name string) {

}
