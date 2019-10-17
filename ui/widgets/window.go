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
	"github.com/rivo/tview"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/models"
)

type Window struct {
	app    *tview.Application
	window *tview.Grid

	// Widgets
	navBar  *NavBar
	status  *status
	browser *Browser
}

func NewWindow() Window {
	w := Window{
		app:     tview.NewApplication(),
		navBar:  NewNavBar(),
		status:  newStatus(),
		window:  tview.NewGrid(),
		browser: NewBrowser(),
	}
	w.window.SetTitle(" " + config.AppName + " ")
	w.window.SetTitleColor(config.ColorPrimary)
	w.window.SetBackgroundColor(config.ColorBackground)
	w.window.SetBorderColor(config.ColroMainFrame)
	w.setLayout()
	w.app.SetRoot(w.window, true)

	w.browser.setData(testData(), models.ArtistList)

	return w
}

func (w *Window) setLayout() {
	w.window.SetBorder(true)
	w.window.SetRows(1, -1, 4)
	w.window.SetColumns(-1)

	w.window.AddItem(w.navBar, 0, 0, 1, 1, 1, 30, false)
	w.window.AddItem(w.browser, 1, 0, 1, 1, 15, 10, false)
	w.window.AddItem(w.status, 2, 0, 1, 1, 3, 10, false)
}

func (w *Window) Run() {
	w.app.Run()

}
