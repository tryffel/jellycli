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
	"github.com/sirupsen/logrus"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
)

// all operations that are callable from context menus
type contextOperator interface {
	AddSongToPlaylist(song *models.Song) error
	ViewAlbumArtist(album *models.Album)
	ViewSongArtist(song *models.Song)
	ViewSongAlbum(song *models.Song)
	ViewArtist(artist *models.Artist)
	InstantMix(item models.Item)
	OpenInBrowser(item models.Item)
}

func (w *Window) AddSongToPlaylist(song *models.Song) error {
	return nil
}

func (w *Window) ViewAlbumArtist(album *models.Album) {
	w.selectAlbum(album)
}

func (w *Window) ViewArtist(artist *models.Artist) {
	w.selectArtist(artist)
}

func (w *Window) ViewSongArtist(song *models.Song) {
	_, artist, err := w.mediaItems.GetSongArtistAlbum(song)
	if err != nil {
		logrus.Errorf("View song album: %v", err)
		return
	}
	w.selectArtist(artist)
}

func (w *Window) ViewSongAlbum(song *models.Song) {
	album, _, err := w.mediaItems.GetSongArtistAlbum(song)
	if err != nil {
		logrus.Errorf("View song album: %v", err)
		return
	}
	w.selectAlbum(album)
}

func (w *Window) InstantMix(item models.Item) {
	if item == nil {
		logrus.Warning("get instant mix on empty item")
		return
	}

	songs, err := w.mediaItems.GetInstantMix(item)
	if err != nil {
		logrus.Errorf("get instant mix: %v", err)
		return
	}

	w.mediaPlayer.StopMedia()
	w.mediaQueue.ClearQueue(true)
	w.mediaQueue.AddSongs(songs)
}

func (w *Window) OpenInBrowser(item models.Item) {
	url := w.mediaItems.GetLink(item)
	if url != "" {
		util.OpenUrlInBrowser(url)
	}
}
