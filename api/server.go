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

package api

import (
	"io"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

//MediaServer uses Jellyfin servers resources and exposes them
type MediaServer interface {
	//GetSongDirect downloads song and returns readcloser if any.
	GetSongDirect(id string, codec string) (io.ReadCloser, error)
	//Search returns Songs that match query
	//TODO: set single / multiple of artist, album, playlist, song
	Search(query string, limit int) (*SearchResult, error)
	//ReportProgress reports current playing progress to server
	ReportProgress(state *interfaces.PlaybackState) error

	//GetItem retrieves single item by its id
	GetItem(id models.Id) (models.Item, error)
	//GetItems retrieves multiple items by their id's
	GetItems(ids []models.Id) ([]models.Item, error)

	//GetArtist gets artist by id.
	GetArtist(id models.Id) (models.Artist, error)

	//Getalbum retrieves album by id.
	GetAlbum(id models.Id) (models.Album, error)
}
