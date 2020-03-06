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
	"fmt"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/jellycli/models"
)

type mediaItemType string

func (m mediaItemType) String() string {
	return string(m)
}

const (
	mediaTypeAlbum        mediaItemType = "MusicAlbum"
	mediaTypeArtist       mediaItemType = "MusicArtist"
	mediaTypeSong         mediaItemType = "Audio"
	mediaTypePlaylist     mediaItemType = "Playlist"
	folderTypePlaylists   mediaItemType = "PlaylistsFolder"
	folderTypeCollections mediaItemType = "CollectionFolder"
	fodlerTypeUserView    mediaItemType = "UserView"
)

// itemType: each item provided by api has Type-field. This interface returns expected type and actual type
type itemType interface {
	// what type
	ExpectType() mediaItemType
	GotType() mediaItemType
}

// assert type matches expected
func assertType(item itemType) error {
	got := item.GotType()
	expect := item.ExpectType()
	if got == expect {
		return nil
	}

	return fmt.Errorf("expect '%s', got '%s'", expect, got)
}

// make type assertion and log failures.
// action is user for logging.
func logInvalidType(item itemType, action string) {
	err := assertType(item)
	if err != nil {
		logrus.Errorf("type error (%s): %v", action, err)
	}
}

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

func (a *artist) ExpectType() mediaItemType {
	return mediaTypeArtist
}

func (a *artist) GotType() mediaItemType {
	return mediaItemType(a.Type)
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

func (a *album) ExpectType() mediaItemType {
	return mediaTypeAlbum
}

func (a *album) GotType() mediaItemType {
	return mediaItemType(a.Type)
}

func (a *album) toAlbum() *models.Album {
	var artist models.Id
	if len(a.Artists) >= 1 {
		artist = models.Id(a.Artists[0].Id)
	}

	artists := make([]models.IdName, len(a.Artists))
	for i, v := range a.Artists {
		artists[i].Name = v.Name
		artists[i].Id = models.Id(v.Id)
	}

	return &models.Album{
		Id:                models.Id(a.Id),
		Name:              a.Name,
		Year:              a.Year,
		Duration:          a.Duration / ticksToSecond,
		Artist:            artist,
		Songs:             nil,
		SongCount:         -1,
		ImageId:           a.ImageTags.Primary,
		DiscCount:         0,
		AdditionalArtists: artists,
	}
}

type songs struct {
	Songs      []song `json:"Items"`
	TotalSongs int    `json:"TotalRecordCount"`
}

type song struct {
	Name           string   `json:"Name"`
	Id             string   `json:"Id"`
	Duration       int      `json:"RunTimeTicks"`
	ProductionYear int      `json:"ProductionYear"`
	IndexNumber    int      `json:"IndexNumber"`
	Type           string   `json:"Type"`
	AlbumId        string   `json:"AlbumId"`
	Album          string   `json:"Album"`
	DiscNumber     int      `json:"ParentIndexNumber"`
	Artists        []nameId `json:"ArtistItems"`
}

func (s *song) ExpectType() mediaItemType {
	return mediaTypeSong
}

func (s *song) GotType() mediaItemType {
	return mediaItemType(s.Type)
}

func (s *song) toSong() *models.Song {
	artists := make([]models.IdName, len(s.Artists))
	for i, v := range s.Artists {
		artists[i].Name = v.Name
		artists[i].Id = models.Id(v.Id)
	}

	return &models.Song{
		Id:         models.Id(s.Id),
		Name:       s.Name,
		Duration:   s.Duration / ticksToSecond,
		Album:      models.Id(s.AlbumId),
		Index:      s.IndexNumber,
		DiscNumber: s.DiscNumber,
		Artists:    artists,
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

type playlists struct {
	Playlists []playlist `json:"Items"`
}

type playlist struct {
	Name     string   `json:"Name"`
	Id       string   `json:"Id"`
	Genres   []string `json:"Genres"`
	Duration int      `json:"RunTimeTicks"`
	Type     string   `json:"Type"`
	Songs    int      `json:"ChildCount"`
}

func (p *playlist) ExpectType() mediaItemType {
	return mediaTypePlaylist
}

func (p *playlist) GotType() mediaItemType {
	return mediaItemType(p.Type)
}

func (p *playlist) toPlaylist() *models.Playlist {
	return &models.Playlist{
		Id:        models.Id(p.Id),
		Name:      p.Name,
		Duration:  p.Duration / ticksToSecond,
		Songs:     nil,
		SongCount: p.Songs,
	}
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
	Primary string `json:"Primary"`
}
