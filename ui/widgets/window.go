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
	"tryffel.net/pkg/jellycli/ui/widgets/modal"
)

type Window struct {
	app    *tview.Application
	window *tview.Grid

	// Widgets
	navBar   *NavBar
	status   *Status
	mediaNav *MediaNavigation
	search   *modal.Search
	view     *modal.ViewModal
	help     *modal.Help
	queue    *modal.Queue
	history  *modal.Queue

	artist *ArtistView
	album  *AlbumView

	gridAxisX  []int
	gridAxisY  []int
	customGrid bool
	modal      modal.Modal

	mediaController controller.MediaController

	hasModal  bool
	lastFocus tview.Primitive
}

func NewWindow(mc controller.MediaController) Window {
	w := Window{
		app:    tview.NewApplication(),
		status: newStatus(mc),
		window: tview.NewGrid(),
		artist: NewArtistView(),
		album:  NewAlbumview(),
	}

	w.mediaNav = NewMediaNavigation(w.selectMedia)
	w.navBar = NewNavBar(w.keyHandlerCb)
	w.mediaController = mc

	w.window.SetTitle(" " + config.AppName + " ")
	w.window.SetTitleColor(config.ColorPrimary)
	w.window.SetBackgroundColor(config.ColorBackground)
	w.window.SetBorderColor(config.ColroMainFrame)
	w.setLayout()
	w.app.SetRoot(w.window, true)
	w.app.SetFocus(w.window)

	w.app.SetInputCapture(w.eventHandler)
	//w.window.SetInputCapture(w.eventHandler)
	w.search = modal.NewSearch(w.searchCb)
	w.search.SetDoneFunc(w.wrapCloseModal(w.search))
	w.help = modal.NewHelp(w.closeHelp)
	w.help.SetDoneFunc(w.wrapCloseModal(w.help))
	w.queue = modal.NewQueue(modal.QueueModeQueue)
	w.queue.SetDoneFunc(w.wrapCloseModal(w.queue))
	w.view = modal.NewViewModal()
	w.view.SetDoneFunc(w.wrapCloseModal(w.view))
	w.view.SetViewFunc(w.openViewFunc)

	w.history = modal.NewQueue(modal.QueueModeHistory)
	w.history.SetDoneFunc(w.wrapCloseModal(w.history))

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
		Duration: 185,
		Artist:   "Artist A",
		ArtistId: "a1",
		Album:    "Album B",
		AlbumId:  "ab2",
		Year:     2019,
	})

	w.mediaController.SetItemsCallback(w.itemsCb)
	w.mediaController.SetStatusCallback(w.statusCb)

	return w
}

func (w *Window) Run() error {
	return w.app.Run()
}

func (w *Window) Stop() {
	w.app.Stop()
}

func (w *Window) setLayout() {
	w.gridAxisY = []int{1, -1, -2, -2, -1, 4}
	w.gridAxisX = []int{24, -1, -2, -2, -1, 24}
	w.window.Clear()
	w.window.SetBorder(true)
	w.window.SetRows(w.gridAxisY...)
	w.window.SetColumns(w.gridAxisX...)

	w.window.AddItem(w.navBar, 0, 0, 1, 6, 1, 30, false)
	w.window.AddItem(w.mediaNav, 1, 0, 4, 1, 5, 10, true)
	w.window.AddItem(w.status, 5, 0, 1, 6, 3, 10, false)
}

func (w *Window) setViewWidget(p tview.Primitive) {
	w.window.AddItem(p, 1, 1, 4, 5, 15, 10, false)
	w.lastFocus = w.app.GetFocus()
	w.app.SetFocus(p)
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
		if w.status.state.State == player.Pause {
			go w.mediaController.Continue()
		} else if w.status.state.State == player.Play {
			go w.mediaController.Pause()
		}
	case ctrls.VolumeDown:
		volume := w.status.state.Volume - 5
		go w.mediaController.SetVolume(volume)
	case ctrls.VolumeUp:
		volume := w.status.state.Volume + 5
		go w.mediaController.SetVolume(volume)
	default:
		return false
	}
	//w.status.InputHandler()(event, nil)
	return true
}

// Open view
func (w *Window) openViewFunc(view controller.View) {
	go w.mediaController.GetView(view)
}

func (w *Window) navBarCtrl(key tcell.Key) bool {
	navBar := config.KeyBinds.NavigationBar
	switch key {
	// Navigation bar
	case navBar.Quit:
		w.app.Stop()
	case navBar.Help:
		w.showModal(w.help, 25, 50, true)
	case navBar.View:
		w.showModal(w.view, 20, 50, true)
	case navBar.Search:
		w.showModal(w.search, 10, 50, true)
	case navBar.Queue:
		w.showModal(w.queue, 20, 60, true)
		w.queue.SetData(w.mediaController.GetQueue(), w.mediaController.QueueDuration())
	case navBar.History:
		w.showModal(w.history, 20, 60, true)
		w.history.SetData(w.mediaController.GetHistory(100), 0)
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
	w.app.SetFocus(w.window)

	if doSearch {
		//w.mediaController.Search(query)
	}

}

func (w *Window) closeHelp() {
	w.app.SetFocus(w.window)
}

func (w *Window) wrapCloseModal(modal modal.Modal) func() {
	return func() {
		w.closeModal(modal)
	}
}

func (w *Window) closeModal(modal modal.Modal) {
	if w.hasModal {
		modal.Blur()
		modal.SetVisible(false)

		modal.SetVisible(false)
		w.window.RemoveItem(modal)
		w.hasModal = false
		w.modal = nil
		if w.customGrid {
			w.window.SetRows(w.gridAxisY...)
			w.window.SetColumns(w.gridAxisX...)
			w.customGrid = false
		}

		w.app.SetFocus(w.lastFocus)
		w.lastFocus = nil
		w.hasModal = false
	} else {
		logrus.Warning("Trying to close modal when there's no open modal.")
	}
}

func (w *Window) showModal(modal modal.Modal, height, width uint, lockSize bool) {
	if !w.hasModal {
		w.hasModal = true
		w.modal = modal
		w.lastFocus = w.app.GetFocus()
		w.lastFocus.Blur()
		if !lockSize {
			w.customGrid = false
			w.window.AddItem(modal, 2, 2, 2, 2, 8, 30, true)
		} else {
			w.customGrid = true
			x := make([]int, len(w.gridAxisX))
			y := make([]int, len(w.gridAxisY))
			copy(x, w.gridAxisX)
			copy(y, w.gridAxisY)
			x[2] = int(width / 2)
			x[3] = x[2]
			y[2] = int(height / 2)
			y[3] = y[2]
			w.window.SetRows(y...)
			w.window.SetColumns(x...)
			w.window.AddItem(modal, 2, 2, 2, 2, int(height), int(width), true)
		}
		w.app.SetFocus(modal)
		modal.SetVisible(true)
		w.app.QueueUpdateDraw(func() {})
	} else {
		logrus.Warning("Trying show close modal when there's another modal open.")
	}
}

func (w *Window) statusCb(state player.PlayingState) {
	w.status.UpdateState(state, nil)
	w.app.QueueUpdateDraw(func() {})
}

func (w *Window) itemsCb(items []models.Item) {
	//w.browser.setData(items)
	w.app.QueueUpdateDraw(func() {})
}

func (w *Window) InitBrowser(items []models.Item) {
	//w.browser.setData(items)
	w.app.Draw()
}

func (w *Window) selectMedia(m MediaSelect) {
	switch m {
	case MediaFavoriteArtists:
		artists, err := w.mediaController.GetFavoriteArtists()
		if err != nil {
			logrus.Errorf("get favorite artists: %v", err)
		} else {
			w.artist.SetArtist(artists[0])
			w.setViewWidget(w.artist)
		}
	}
}
