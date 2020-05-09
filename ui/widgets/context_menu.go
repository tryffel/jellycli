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

package widgets

import (
	"github.com/sirupsen/logrus"
	"tryffel.net/go/jellycli/models"
)

// all operations that are callable from context menus
type contextOperator interface {
	AddSongToPlaylist(song *models.Song) error
	ViewAlbumArtist(album *models.Album)
	ViewSongArtist(song *models.Song)
	ViewSongAlbum(song *models.Song)
	ViewArtist(artist *models.Artist)
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
