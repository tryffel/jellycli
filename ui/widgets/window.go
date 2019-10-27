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
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/controller"
	"tryffel.net/pkg/jellycli/models"
	"tryffel.net/pkg/jellycli/player"
)

type Window struct {
	app    *tview.Application
	window *tview.Grid

	// Widgets
	navBar  *NavBar
	status  *Status
	browser *Browser
	search  *Search
	help    *Help

	mediaController controller.MediaController

	hasModal  bool
	lastFocus tview.Primitive
}

func NewWindow(mc controller.MediaController) Window {
	w := Window{
		app:    tview.NewApplication(),
		status: newStatus(mc),
		window: tview.NewGrid(),
	}

	w.navBar = NewNavBar(w.keyHandlerCb)
	w.mediaController = mc
	w.browser = NewBrowser(mc)

	w.window.SetTitle(" " + config.AppName + " ")
	w.window.SetTitleColor(config.ColorPrimary)
	w.window.SetBackgroundColor(config.ColorBackground)
	w.window.SetBorderColor(config.ColroMainFrame)
	w.setLayout()
	w.app.SetRoot(w.window, true)
	w.app.SetFocus(w.browser)

	//data := testData()
	//w.browser.setData(&data, models.TypeArtist)

	//w.browser.SetInitialData(mc.)

	w.app.SetInputCapture(w.eventHandler)
	//w.window.SetInputCapture(w.eventHandler)
	w.search = NewSearch(w.searchCb)
	w.help = NewHelp(w.closeHelp)

	w.status.UpdateState(player.PlayingState{
		State:               player.Play,
		PlayingType:         player.Song,
		Song:                "Song",
		Artist:              "Artist",
		Album:               "Album",
		CurrentSongDuration: 185,
		CurrentSongPast:     92,
		PlaylistDuration:    0,
		PlaylistLeft:        0,
		Volume:              50,
	}, &models.SongInfo{
		Id:       "song1",
		Name:     "Song",
		Length:   185,
		Artist:   "Artist A",
		ArtistId: "a1",
		Album:    "Album B",
		AlbumId:  "ab2",
		Year:     2019,
	})

	return w
}

func (w *Window) Run() error {
	return w.app.Run()
}

func (w *Window) Stop() {
	w.app.Stop()
}

func (w *Window) setLayout() {
	w.window.Clear()
	w.window.SetBorder(true)
	w.window.SetRows(1, -1, 4)
	w.window.SetColumns(-1)

	w.window.AddItem(w.navBar, 0, 0, 1, 1, 1, 30, false)
	w.window.AddItem(w.browser, 1, 0, 1, 1, 15, 10, false)
	w.window.AddItem(w.status, 2, 0, 1, 1, 3, 10, false)
}

func (w *Window) eventHandler(event *tcell.EventKey) *tcell.EventKey {

	out := w.keyHandler(event)
	if out == nil {
		return nil
	}
	return event
}

// Keyhandler that has to react to buttons or drop them completely
func (w *Window) keyHandlerCb(key *tcell.Key) {

}

// Key handler, if match, return nil
func (w *Window) keyHandler(event *tcell.EventKey) *tcell.Key {
	key := event.Key()
	/*
		if key >= tcell.KeyF1 && key <= tcell.KeyF12 && !w.navBarFocused{
			//Activate navigation bar on function button
			w.lastFocus = w.app.GetFocus()
			w.lastFocus.Blur()
			w.app.SetFocus(w.navBar)
			w.navBarFocused = true
		} else if key == tcell.KeyEscape && w.navBarFocused {
			//Deactivate navigation bar and return to last focus
			w.navBarFocused = false
			w.navBar.Blur()
			w.app.SetFocus(w.lastFocus)
			w.lastFocus = nil
			return nil
		}
	*/

	if w.mediaCtrl(event) {
		return nil
	}
	if w.navBarCtrl(key) {
		return nil
	}
	if w.moveCtrl(key) {
		return nil
	}
	// Moving around
	return &key
}

func (w *Window) mediaCtrl(event *tcell.EventKey) bool {
	ctrls := config.KeyBinds.Global
	key := event.Key()
	switch key {
	case ctrls.PlayPause:
		break
	case ctrls.VolumeDown:
		break
	case ctrls.VolumeUp:
		break
	default:
		return false
	}
	w.status.InputHandler()(event, nil)
	return true
}

func (w *Window) navBarCtrl(key tcell.Key) bool {
	navBar := config.KeyBinds.NavigationBar
	switch key {
	// Navigation bar
	case navBar.Quit:
		w.app.Stop()
	case navBar.Search:
		if !w.hasModal {
			w.hasModal = true
			w.lastFocus = w.app.GetFocus()
			w.lastFocus.Blur()
			w.browser.AddModal(w.search, 10, 50, true)
			w.app.SetFocus(w.search)
		}
	case navBar.Help:
		if !w.hasModal {
			w.hasModal = true
			w.lastFocus = w.app.GetFocus()
			w.lastFocus.Blur()
			w.browser.AddModal(w.help, 25, 50, true)
			w.app.SetFocus(w.help)
		}
	case tcell.KeyEscape:
		if w.hasModal {
			w.hasModal = false
			w.browser.Blur()
			w.lastFocus.Focus(nil)
			w.lastFocus = nil
			return false
		}
	default:
		return false
	}
	return true
}

func (w *Window) moveCtrl(key tcell.Key) bool {
	return false
}

func (w *Window) searchCb(query string, doSearch bool) {
	logrus.Debug("In search callback")
	w.browser.RemoveModal(w.search)
	w.app.SetFocus(w.window)

	if doSearch {
		//w.mediaController.Search(query)
	}

}

func (w *Window) closeHelp() {
	w.browser.RemoveModal(w.help)
	w.app.SetFocus(w.window)

}

func (w *Window) statusCb(state player.PlayingState) {
	w.status.UpdateState(state, nil)
}

func (w *Window) itemsCb(items []models.Item) {
	w.browser.setData(items)
}

func (w *Window) InitBrowser(items []models.Item) {
	w.browser.setData(items)

}
