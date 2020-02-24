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
	"github.com/sirupsen/logrus"
	"tryffel.net/go/jellycli/models"
)

const (
	mediaTypeAlbum        = "MusicAlbum"
	mediaTypeArtist       = "MusicArtist"
	mediaTypeSong         = "Audio"
	mediaTypePlaylist     = "Playlist"
	folderTypePlaylists   = "PlaylistsFolder"
	folderTypeCollections = "CollectionFolder"
	fodlerTypeUserView    = "UserView"
)

type nameId struct {
	Name string `json:"Name"`
	Id   string `json:"Id"`
}

type artists struct {
	Artists      []artist `json:"Items"`
	TotalArtists int      `json:"TotalRecordCount"`
}

type artist struct {
	Name          string `json:"Name"`
	Id            string `json:"Id"`
	TotalDuration int    `json:"RunTimeTicks"`
	Type          string `json:"Type"`
	TotalSongs    int    `json:"SongCount"`
	TotalAlbums   int    `json:"AlbumCount"`
}

func (a *artist) toArtist() *models.Artist {
	return &models.Artist{
		Id:            models.Id(a.Id),
		Name:          a.Name,
		Albums:        nil,
		TotalDuration: a.TotalDuration / ticksToSecond,
		AlbumCount:    a.TotalAlbums,
	}
}

type albums struct {
	Albums      []album `json:"Items"`
	TotalAlbums int     `json:"TotalRecordCount"`
}

type album struct {
	Name      string   `json:"Name"`
	Id        string   `json:"Id"`
	Duration  int      `json:"RunTimeTicks"`
	Year      int      `json:"ProductionYear"`
	Type      string   `json:"Type"`
	Artists   []nameId `json:"AlbumArtists"`
	Overview  string   `json:"Overview"`
	Genres    []string `json:"Genres"`
	ImageTags images   `json:"ImageTags"`
}

func (a *album) toAlbum() *models.Album {
	var artist models.Id
	if len(a.Artists) >= 1 {
		artist = models.Id(a.Artists[0].Id)
	}

	return &models.Album{
		Id:        models.Id(a.Id),
		Name:      a.Name,
		Year:      a.Year,
		Duration:  a.Duration / ticksToSecond,
		Artist:    artist,
		Songs:     nil,
		SongCount: -1,
		ImageId:   a.ImageTags.Primary,
	}
}

type songs struct {
	Songs      []song `json:"Items"`
	TotalSongs int    `json:"TotalRecordCount"`
}

type song struct {
	Name           string `json:"Name"`
	Id             string `json:"Id"`
	Duration       int    `json:"RunTimeTicks"`
	ProductionYear int    `json:"ProductionYear"`
	IndexNumber    int    `json:"IndexNumber"`
	Type           string `json:"Type"`
	AlbumId        string `json:"AlbumId"`
	Album          string `json:"Album"`
}

func (s *song) toSong() *models.Song {
	if s.Type != mediaTypeSong {
		logrus.Warningf("Converting (%s) to song", s.Type)
	}
	return &models.Song{
		Id:       models.Id(s.Id),
		Name:     s.Name,
		Duration: s.Duration / ticksToSecond,
		Album:    models.Id(s.AlbumId),
		Index:    s.IndexNumber,
	}
}

type collections struct {
	Collections []collection `json:"Items"`
}

type collection struct {
	Name           string `json:"Name"`
	Id             string `json:"Id"`
	Type           string `json:"Type"`
	CollectionType string `json:"CollectionType"`
}

type playlist struct {
	Name     string   `json:"Name"`
	Id       string   `json:"Id"`
	Genres   []string `json:"Genres"`
	Duration int      `json:"RunTimeTicks"`
	Type     string   `json:"Type"`
	Songs    int      `json:"ChildCound"`
}

type view struct {
	nameId
	Type string `json:"Type"`
}

func (v *view) toView() *models.View {
	return &models.View{
		Name: v.Name,
		Id:   models.Id(v.Id),
		Type: v.Type,
	}
}

type views struct {
	Views []view `json:"Items"`
}

type images struct {
	Primary string `json:"Primary`
}
