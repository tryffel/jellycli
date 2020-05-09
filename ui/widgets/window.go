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

// Package widgets contains all widgets that are used in jellycli. Window is the root widget and controls access
// to interfaces.Player, interfaces.Queue and interfaces.Items.
package widgets

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/ui/widgets/modal"
	"tryffel.net/go/twidgets"
)

type Window struct {
	app    *cview.Application
	layout *twidgets.ModalLayout

	// Widgets
	navBar   *twidgets.NavBar
	status   *Status
	mediaNav *MediaNavigation
	help     *modal.Help
	message  *modal.Message
	queue    *Queue
	history  *History

	albumList     *AlbumList
	similarAlbums *AlbumList
	album         *AlbumView
	artistList    *ArtistList
	playlists     *Playlists
	playlist      *PlaylistView
	songs         *SongList
	genres        *GenreList

	gridAxisX  []int
	gridAxisY  []int
	customGrid bool
	modal      modal.Modal

	mediaView         Previous
	mediaViewSelected bool

	mediaPlayer interfaces.Player
	mediaItems  interfaces.ItemController
	mediaQueue  interfaces.QueueController

	hasModal  bool
	lastFocus cview.Primitive
}

func NewWindow(p interfaces.Player, i interfaces.ItemController, q interfaces.QueueController) Window {
	w := Window{
		app:    cview.NewApplication(),
		status: newStatus(p),
		layout: twidgets.NewModalLayout(),
	}

	w.artistList = NewArtistList(w.selectArtist)
	w.artistList.SetBackCallback(w.goBack)
	w.artistList.selectPageFunc = w.showArtistPage
	w.albumList = NewAlbumList(w.selectAlbum)
	w.albumList.SetBackCallback(w.goBack)
	w.albumList.selectPageFunc = w.showAlbumPage
	w.albumList.similarFunc = w.showSimilarArtists

	w.similarAlbums = NewAlbumList(w.selectAlbum)
	w.similarAlbums.SetBackCallback(w.goBack)
	w.similarAlbums.EnablePaging(false)

	w.album = NewAlbumview(w.playSong, w.playSongs, &w)
	w.album.SetBackCallback(w.goBack)
	w.album.similarFunc = w.showSimilarAlbums
	w.mediaNav = NewMediaNavigation(w.selectMedia)
	w.navBar = twidgets.NewNavBar(config.Color.NavBar.ToWidgetsNavBar(), w.navBarHandler)

	w.playlists = NewPlaylists(w.selectPlaylist)
	w.playlist = NewPlaylistView(w.playSong, w.playSongs, &w)
	w.playlist.SetBackCallback(w.goBack)

	w.genres = NewGenreList()
	w.genres.SetBackCallback(w.goBack)
	w.genres.selectFunc = w.selectGenre
	w.genres.selectPageFunc = w.showGenrePage

	w.songs = NewSongList(w.playSong, w.playSongs, &w)
	w.songs.SetBackCallback(w.goBack)
	w.songs.showPage = w.selectSongs
	w.mediaPlayer = p
	w.mediaItems = i
	w.mediaQueue = q

	w.setLayout()
	w.app.SetRoot(w.layout, true)
	if config.AppConfig.Player.MouseEnabled {
		w.app.EnableMouse(true)
		interval := time.Millisecond * time.Duration(config.AppConfig.Player.DoubleClickMs)
		w.app.SetDoubleClickInterval(interval)
	}

	w.app.SetFocus(w.mediaNav)

	w.app.SetInputCapture(w.eventHandler)
	//w.window.SetInputCapture(w.eventHandler)
	w.help = modal.NewHelp(w.closeHelp)
	w.help.SetDoneFunc(w.wrapCloseModal(w.help))
	w.message = modal.NewMessage()
	w.message.SetDoneFunc(w.closeMessage)

	w.queue = NewQueue()
	w.queue.SetBackCallback(w.goBack)
	w.queue.clearFunc = w.clearQueue
	w.queue.controller = w.mediaQueue
	w.mediaQueue.AddQueueChangedCallback(func(songs []*models.Song) {
		w.app.QueueUpdateDraw(func() {
			index := w.queue.list.GetSelectedIndex()
			w.queue.SetSongs(songs)
			w.queue.list.SetSelected(index)
		})
	})

	w.history = NewHistory()
	w.history.SetBackCallback(w.goBack)

	w.mediaQueue.SetHistoryChangedCallback(func(songs []*models.Song) {
		w.app.QueueUpdateDraw(func() {
			w.history.SetSongs(songs)
		})
	})

	w.layout.Grid().SetBackgroundColor(config.Color.Background)

	w.mediaPlayer.AddStatusCallback(w.statusCb)

	navBarLabels := []string{"Help", "Queue", "History"}

	sc := config.KeyBinds.NavigationBar
	navBarShortucts := []tcell.Key{sc.Help, sc.Queue, sc.History}

	for i, v := range navBarLabels {
		btn := cview.NewButton(v)
		w.navBar.AddButton(btn, navBarShortucts[i])
	}

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

	w.layout.SetGridXSize([]int{10, -1, -1, -1, -1, -1, -1, -1, -1, 10})
	w.layout.SetGridYSize([]int{1, -1, -1, -1, -1, -1, -1, -1, -1, 5})

	w.layout.Grid().AddItem(w.navBar, 0, 0, 1, 10, 1, 30, false)
	w.layout.Grid().AddItem(w.mediaNav, 1, 0, 8, 2, 5, 10, false)
	w.layout.Grid().AddItem(w.status, 9, 0, 1, 10, 3, 10, false)

	//w.setViewWidget(w.artistList)
}

// go back to previous primitive
func (w *Window) goBack(p Previous) {
	w.setViewWidget(p, false)
}

// set central widget. If updatePrevious, set update previous primitive's last primitive
func (w *Window) setViewWidget(p Previous, updatePrevious bool) {
	if p == w.mediaView {
		return
	}

	last := w.mediaView
	w.lastFocus = w.app.GetFocus()
	w.layout.Grid().RemoveItem(w.mediaView)
	w.layout.Grid().AddItem(p, 1, 2, 8, 8, 15, 10, false)
	w.app.SetFocus(p)
	w.mediaView = p
	if updatePrevious {
		p.SetLast(last)
	}
}

func (w *Window) eventHandler(event *tcell.EventKey) *tcell.EventKey {

	out := w.keyHandler(event)
	if out == nil {
		return nil
	}
	return event
}

func (w *Window) navBarHandler(label string) {

}

// Key handler, if match, return nil
func (w *Window) keyHandler(event *tcell.EventKey) *tcell.Key {

	key := event.Key()
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
	case ctrls.Stop:
		w.mediaPlayer.StopMedia()
		w.mediaQueue.ClearQueue(true)
	case ctrls.PlayPause:
		w.mediaPlayer.PlayPause()
	case ctrls.VolumeDown:
		volume := w.status.state.Volume.Add(-5)
		go w.mediaPlayer.SetVolume(volume)
	case ctrls.VolumeUp:
		volume := w.status.state.Volume.Add(5)
		go w.mediaPlayer.SetVolume(volume)
	case ctrls.Next:
		w.mediaPlayer.Next()
	case ctrls.Previous:
		w.mediaPlayer.Previous()
	case ctrls.Forward:
		w.mediaPlayer.Seek(interfaces.AudioTick(3000))
	case ctrls.Backward:
		w.mediaPlayer.Seek(interfaces.AudioTick(-3000))
	default:
		return false
	}
	return true
}

func (w *Window) navBarCtrl(key tcell.Key) bool {
	navBar := config.KeyBinds.NavigationBar
	switch key {
	// Navigation bar
	case navBar.Quit:
		w.app.Stop()
	case navBar.Help:
		stats := w.mediaItems.GetStatistics()
		w.help.SetStats(stats)
		w.showModal(w.help, 25, 50, true)
	case navBar.Queue:
		if w.help.HasFocus() {
			w.closeModal(w.help)
		}
		w.setViewWidget(w.queue, true)
	case navBar.History:
		if w.help.HasFocus() {
			w.closeModal(w.help)
		}
		w.setViewWidget(w.history, true)
		items := w.mediaQueue.GetHistory(100)
		duration := 0
		for _, v := range items {
			duration += v.Duration
		}
	default:
		return false
	}
	return true
}

func (w *Window) moveCtrl(key tcell.Key) bool {
	if key == tcell.KeyTAB {
		if w.mediaViewSelected {
			w.lastFocus = w.mediaView
			w.app.SetFocus(w.mediaNav)
			if w.lastFocus != nil {
				w.lastFocus.Blur()
			}
			w.mediaViewSelected = false
		} else {
			w.lastFocus = w.app.GetFocus()
			w.app.SetFocus(w.mediaView)
			w.mediaViewSelected = true
			if w.lastFocus != nil {
				w.lastFocus.Blur()
			}
		}
		return true
	}
	return false
}

func (w *Window) searchCb(query string, doSearch bool) {
	logrus.Debug("In search callback")
	w.app.SetFocus(w.layout)

	if doSearch {
		//w.mediaController.Search(query)
	}

}

func (w *Window) closeHelp() {
	w.app.SetFocus(w.layout)
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
		w.layout.RemoveModal(modal)

		w.hasModal = false
		w.modal = nil
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
			w.layout.AddFixedModal(modal, height, width, twidgets.ModalSizeMedium)
		} else {
			w.layout.AddDynamicModal(modal, twidgets.ModalSizeLarge)
		}
		w.app.SetFocus(modal)
		modal.SetVisible(true)
		w.app.QueueUpdateDraw(func() {})
	} else {
		logrus.Warning("Trying show close modal when there's another modal open.")
	}
}

func (w *Window) statusCb(state interfaces.AudioStatus) {
	w.status.UpdateState(state, nil)
	w.app.QueueUpdateDraw(func() {})
}

func (w *Window) InitBrowser(items []models.Item) {
	w.app.Draw()
}

func (w *Window) selectMedia(m MediaSelect) {
	switch m {
	case MediaLatestMusic:
		albums, err := w.mediaItems.GetLatestAlbums()
		if err != nil {
			logrus.Errorf("get favorite artists: %v", err)
		} else {
			duration := 0
			for _, v := range albums {
				duration += v.Duration
			}
			// set pseudo artist
			artist := &models.Artist{
				Id:            "",
				Name:          "Latest albums",
				Albums:        nil,
				TotalDuration: duration,
				AlbumCount:    len(albums),
			}

			w.mediaNav.SetCount(MediaLatestMusic, len(albums))
			w.albumList.Clear()
			w.albumList.EnablePaging(false)
			w.albumList.EnableSimilar(false)
			w.albumList.EnableArtistMode(false)
			w.albumList.SetArtist(artist)
			w.albumList.SetAlbums(albums)
			w.setViewWidget(w.albumList, true)
		}
	case MediaFavoriteArtists:
		artists, err := w.mediaItems.GetFavoriteArtists()
		if err != nil {
			logrus.Errorf("get favorite artists: %v", err)
		} else {
			w.artistList.Clear()
			w.artistList.SetText("Favorite artists")
			w.artistList.EnablePaging(false)
			w.mediaNav.SetCount(MediaFavoriteArtists, len(artists))
			w.artistList.AddArtists(artists)
			w.setViewWidget(w.artistList, true)
		}
	case MediaPlaylists:
		playlists, err := w.mediaItems.GetPlaylists()
		if err != nil {
			logrus.Errorf("get playlists: %v", err)
		} else {
			w.mediaNav.SetCount(MediaPlaylists, len(playlists))
			w.playlists.SetPlaylists(playlists)
			w.setViewWidget(w.playlists, true)
		}
	case MediaSongs, MediaRecent:
		page := interfaces.DefaultPaging()
		var songs []*models.Song
		var count int
		var err error

		if m == MediaSongs {
			songs, count, err = w.mediaItems.GetSongs(0, page.PageSize)
			if err != nil {
				logrus.Errorf("get songs: %v", err)
			} else {
				w.songs.showPage = w.selectSongs
				w.mediaNav.SetCount(m, count)
				w.songs.setTitle("All albums")
			}
		} else {
			songs, count, err = w.mediaItems.GetRecentlyPlayed(page)
			if err != nil {
				logrus.Errorf("get songs: %v", err)
			} else {
				w.songs.showPage = w.showRecentSongsPage
				w.songs.setTitle("Recently played")
				if !config.LimitRecentlyPlayed {
					w.mediaNav.SetCount(m, count)
				}
			}
		}
		page.SetTotalItems(count)
		w.songs.SetSongs(songs, page)

		w.setViewWidget(w.songs, true)
	case MediaArtists, MediaAlbumArtists:
		paging := interfaces.DefaultPaging()
		var artists []*models.Artist
		var err error
		var total int
		var title string
		if m == MediaArtists {
			title = "All artists"
			artists, total, err = w.mediaItems.GetArtists(paging)
		} else if m == MediaAlbumArtists {
			title = "All album artists"
			artists, total, err = w.mediaItems.GetAlbumArtists(paging)
		}
		if err != nil {
			logrus.Errorf("get all artists: %v", err)
			return
		}
		paging.SetTotalItems(total)
		w.mediaNav.SetCount(m, total)

		w.artistList.Clear()
		w.artistList.EnablePaging(true)
		w.artistList.SetPage(paging)

		w.artistList.AddArtists(artists)
		w.setViewWidget(w.artistList, true)
		w.artistList.SetText(fmt.Sprintf("%s: %d", title, paging.TotalItems))
	case MediaAlbums, MediaFavoriteAlbums:
		paging := interfaces.DefaultPaging()
		var albums []*models.Album
		var err error
		var total int
		var title string

		if m == MediaAlbums {
			albums, total, err = w.mediaItems.GetAlbums(paging)
			title = "All Albums"
			w.albumList.EnablePaging(true)
		} else if m == MediaFavoriteAlbums {
			paging.PageSize = 200
			albums, total, err = w.mediaItems.GetFavoriteAlbums(paging)
			title = "Favorite albums"
			w.albumList.EnablePaging(false)
		}

		if err != nil {
			logrus.Errorf("get %s: %v", title, err)
			return
		}
		paging.SetTotalItems(total)
		w.mediaNav.SetCount(m, total)

		w.albumList.SetPage(paging)
		w.albumList.Clear()
		w.albumList.EnableSimilar(false)
		w.albumList.EnableArtistMode(false)

		w.albumList.SetText(fmt.Sprintf("%s\nTotal %v", title, paging.TotalItems))
		w.albumList.SetAlbums(albums)
		w.setViewWidget(w.albumList, true)
	case MediaGenres:
		paging := interfaces.DefaultPaging()
		w.showGenrePage(paging)
	}
}

func (w *Window) selectArtist(artist *models.Artist) {
	albums, err := w.mediaItems.GetArtistAlbums(artist.Id)
	if err != nil {
		logrus.Errorf("get albumList albums: %v", err)
	} else {
		artist.AlbumCount = len(albums)
		w.albumList.Clear()
		w.albumList.EnableArtistMode(true)
		w.albumList.EnablePaging(false)
		w.albumList.EnableSimilar(true)
		w.albumList.SetArtist(artist)
		w.albumList.SetAlbums(albums)
		w.setViewWidget(w.albumList, true)
	}
}

func (w *Window) selectAlbum(album *models.Album) {
	songs, err := w.mediaItems.GetAlbumSongs(album.Id)
	if err != nil {
		logrus.Errorf("get album songs: %v", err)
	} else {
		for _, v := range songs {
			v.AlbumArtist = album.Artist
		}

		w.album.SetAlbum(album, songs)
		w.album.SetLast(w.mediaView)
		w.setViewWidget(w.album, true)
	}
}

func (w *Window) selectPlaylist(playlist *models.Playlist) {
	err := w.mediaItems.GetPlaylistSongs(playlist)
	if err != nil {
		logrus.Warningf("did not get playlist songs: %v", err)
		return
	}

	w.playlist.SetPlaylist(playlist)
	w.setViewWidget(w.playlist, true)
}

func (w *Window) selectSongs(page interfaces.Paging) {
	songs, _, err := w.mediaItems.GetSongs(page.CurrentPage, page.PageSize)
	if err != nil {
		logrus.Errorf("get songs: %v", err)
	}

	w.songs.SetSongs(songs, page)
	w.setViewWidget(w.songs, true)
}

func (w *Window) showRecentSongsPage(page interfaces.Paging) {
	songs, _, err := w.mediaItems.GetRecentlyPlayed(page)
	if err != nil {
		logrus.Errorf("get songs: %v", err)
	}

	w.songs.SetSongs(songs, page)
	w.setViewWidget(w.songs, true)
}

func (w *Window) showArtistPage(page interfaces.Paging) {
	artists, _, err := w.mediaItems.GetArtists(page)
	if err != nil {
		logrus.Errorf("get albumList page: %v", err)
		return
	}

	w.artistList.Clear()
	w.artistList.AddArtists(artists)
	w.artistList.EnablePaging(true)
	w.setViewWidget(w.artistList, false)
}

func (w *Window) showAlbumPage(page interfaces.Paging) {
	albums, total, err := w.mediaItems.GetAlbums(page)
	if err != nil {
		logrus.Errorf("get all albums: %v", err)
		return
	}
	page.SetTotalItems(total)
	w.mediaNav.SetCount(MediaAlbums, total)

	w.albumList.SetPage(page)
	w.albumList.Clear()
	w.albumList.EnablePaging(true)
	w.albumList.SetAlbums(albums)
}

func (w *Window) playSong(song *models.Song) {
	w.playSongs([]*models.Song{song})
}

func (w *Window) playSongs(songs []*models.Song) {
	w.mediaQueue.AddSongs(songs)
}

func (w *Window) clearQueue() {
	w.mediaQueue.ClearQueue(false)
}

func (w *Window) showSimilarArtists(artist models.Id) {
	artists, err := w.mediaItems.GetSimilarArtists(artist)
	if err != nil {
		logrus.Errorf("get similar artists: %v", err)
	} else if len(artists) > 0 {
		w.artistList.Clear()
		w.artistList.AddArtists(artists)
		w.artistList.SetText(fmt.Sprintf("Similar artists: %d", len(artists)))
		w.setViewWidget(w.artistList, true)
	} else {
		w.showMessage("No similar artists", 3, -1, false)
	}
}

func (w *Window) showSimilarAlbums(album *models.Album) {
	albums, err := w.mediaItems.GetSimilarAlbums(album.Id)
	if err != nil {
		logrus.Errorf("get similar artists: %v", err)
	} else if len(albums) > 0 {
		w.similarAlbums.Clear()
		w.similarAlbums.EnableSimilar(false)
		w.similarAlbums.EnablePaging(false)
		w.similarAlbums.SetAlbums(albums)
		w.similarAlbums.SetText(fmt.Sprintf("Similar albums: %d", len(albums)))
		w.setViewWidget(w.similarAlbums, true)
	} else {
		w.showMessage("No similar albums", 3, -1, false)
	}
}

func (w *Window) closeMessage() {
	w.closeModal(w.message)
}

func (w *Window) showMessage(msg string, height, width int, lockSize bool) {
	w.message.SetText(msg)
	if height == -1 {
		height = 25
	}
	if width == -1 {
		width = 50
	}
	w.showModal(w.message, uint(height), uint(width), lockSize)
}

func (w *Window) selectGenre(id models.IdName) {

	albums, err := w.mediaItems.GetGenreAlbums(id)
	if err != nil {
		logrus.Errorf("get genre albums: %v", err)
		return
	}

	w.albumList.Clear()
	w.albumList.EnablePaging(false)
	w.albumList.EnableSimilar(false)
	w.albumList.EnableArtistMode(false)
	w.albumList.SetAlbums(albums)
	w.albumList.SetText("Genre " + id.Name)
	w.setViewWidget(w.albumList, true)
}

func (w *Window) showGenrePage(paging interfaces.Paging) {
	genres, n, err := w.mediaItems.GetGenres(paging)
	if err != nil {
		logrus.Errorf("get genres: %v", err)
		return
	}
	paging.SetTotalItems(n)
	w.genres.SetPage(paging)
	w.genres.setGenres(genres)
	w.setViewWidget(w.genres, true)
}
