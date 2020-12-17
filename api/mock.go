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

package api

import (
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

var MockArtists = []*models.Artist{
	{
		Id:            "artist-1",
		Name:          "artist 1",
		Albums:        []models.Id{"album-1", "album-2"},
		TotalDuration: 3600,
		AlbumCount:    2,
	},
	{
		Id:            "artist-2",
		Name:          "artist 2",
		Albums:        []models.Id{"album-3"},
		TotalDuration: 3600,
		AlbumCount:    1,
	},
}

var MockAlbums = []*models.Album{
	{
		Id:        "album-1",
		Name:      "album-1",
		Year:      2020,
		Duration:  3600,
		Artist:    "artist-1",
		SongCount: 2,
		DiscCount: 1,
		Songs:     []models.Id{"song-1", "song-2"},
	},
	{
		Id:        "album-2",
		Name:      "album-2",
		Year:      2019,
		Duration:  3600,
		Artist:    "artist-1",
		SongCount: 2,
		DiscCount: 1,
		Songs:     []models.Id{"song-3", "song-4"},
	},
	{
		Id:        "album-3",
		Name:      "album-3",
		Year:      2018,
		Duration:  3600,
		Artist:    "artist-2",
		SongCount: 2,
		DiscCount: 1,
		Songs:     []models.Id{"song-5", "song-6"},
	},
}

var MockSongs = []*models.Song{
	{
		Id:          "song-1",
		Name:        "song 1",
		Duration:    180,
		Index:       0,
		Album:       "album-1",
		DiscNumber:  0,
		AlbumArtist: "artist-1",
	},
	{
		Id:          "song-2",
		Name:        "song 2",
		Duration:    180,
		Index:       0,
		Album:       "album-1",
		DiscNumber:  0,
		AlbumArtist: "artist-1",
	},
	{
		Id:          "song-3",
		Name:        "song 3",
		Duration:    180,
		Index:       0,
		Album:       "album-2",
		DiscNumber:  0,
		AlbumArtist: "artist-1",
	},
	{
		Id:          "song-4",
		Name:        "song 4",
		Duration:    180,
		Index:       0,
		Album:       "album-2",
		DiscNumber:  0,
		AlbumArtist: "artist-1",
	},
	{
		Id:          "song-5",
		Name:        "song 5",
		Duration:    180,
		Index:       0,
		Album:       "album-3",
		DiscNumber:  0,
		AlbumArtist: "artist-2",
	},
	{
		Id:          "song-6",
		Name:        "song 6",
		Duration:    180,
		Index:       0,
		Album:       "album-3",
		DiscNumber:  0,
		AlbumArtist: "artist-2",
	},
}

func limitPaging(lastIndex, items int) int {
	if lastIndex > items-1 {
		return items - 1
	}
	return lastIndex
}

type MockConfig struct {
}

func (m *MockConfig) DumpConfig() interface{} {
	return map[string]string{"host": "mock"}
}

func (m *MockConfig) GetType() string {
	return "mock"
}

type MockServer struct {
	Artists         []*models.Artist
	FavoriteArtists []*models.Artist
	AlbumArtists    []*models.Artist
	Albums          []*models.Album
	AlbumSongs      map[models.Id][]*models.Song
	Playlists       []*models.Playlist
	PlaylistSongs   map[models.Id]*models.Song
	Songs           []*models.Song
}

func NewMockServer() *MockServer {
	server := &MockServer{
		Artists:         MockArtists,
		FavoriteArtists: MockArtists,
		AlbumArtists:    MockArtists,
		Albums:          MockAlbums,
		Songs:           MockSongs,
	}
	return server
}

func (m *MockServer) GetArtists(query *interfaces.QueryOpts) ([]*models.Artist, int, error) {
	if query == nil {
		return m.Artists, len(m.Artists), nil
	}

	offset := query.Paging.Offset()
	last := limitPaging(query.Paging.CurrentPage*query.Paging.PageSize, len(m.Artists))
	artists := m.Artists[offset:last]
	return artists, len(m.Artists), nil
}

func (m *MockServer) GetAlbumArtists(query *interfaces.QueryOpts) ([]*models.Artist, int, error) {
	offset := query.Paging.Offset()
	last := limitPaging(query.Paging.CurrentPage*query.Paging.PageSize, len(m.AlbumArtists))
	artists := m.AlbumArtists[offset:last]
	return artists, len(m.Artists), nil
}

func (m *MockServer) GetAlbums(query *interfaces.QueryOpts) ([]*models.Album, int, error) {
	if query == nil {
		return m.Albums, len(m.Albums), nil
	}
	offset := query.Paging.Offset()
	last := limitPaging(query.Paging.CurrentPage*query.Paging.PageSize, len(m.Albums))
	albums := m.Albums[offset:last]
	return albums, len(m.Albums), nil
}

func (m *MockServer) GetArtistAlbums(artist models.Id) ([]*models.Album, error) {
	panic("not implemented")
}

func (m *MockServer) GetAlbumSongs(album models.Id) ([]*models.Song, error) {
	panic("not implemented")
}

func (m *MockServer) GetPlaylists() ([]*models.Playlist, error) {
	panic("not implemented")
}

func (m *MockServer) GetPlaylistSongs(playlist models.Id) ([]*models.Song, error) {
	panic("not implemented")
}

func (m *MockServer) GetFavoriteAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	panic("not implemented")
}

func (m *MockServer) GetSimilarArtists(artist models.Id) ([]*models.Artist, error) {
	panic("not implemented")
}

func (m *MockServer) GetSimilarAlbums(album models.Id) ([]*models.Album, error) {
	panic("not implemented")
}

func (m *MockServer) GetLatestAlbums() ([]*models.Album, error) {
	panic("not implemented")
}

func (m *MockServer) GetRecentlyPlayed(paging interfaces.Paging) ([]*models.Song, int, error) {
	panic("not implemented")
}

func (m *MockServer) GetSongs(query *interfaces.QueryOpts) ([]*models.Song, int, error) {
	panic("not implemented")
}

func (m *MockServer) GetGenres(paging interfaces.Paging) ([]*models.IdName, int, error) {
	panic("not implemented")
}

func (m *MockServer) GetGenreAlbums(genre models.IdName) ([]*models.Album, error) {
	panic("not implemented")
}

func (m *MockServer) GetAlbumArtist(album *models.Album) (*models.Artist, error) {
	panic("not implemented")
}

func (m *MockServer) GetInstantMix(item models.Item) ([]*models.Song, error) {
	panic("not implemented")
}

func (m *MockServer) GetLink(item models.Item) string {
	panic("not implemented")
}

func (m *MockServer) Search(query string, itemType models.ItemType, maxResults int) ([]models.Item, error) {
	panic("not implemented")
}

func (m *MockServer) GetAlbum(id models.Id) (*models.Album, error) {
	panic("not implemented")
}

func (m *MockServer) GetArtist(id models.Id) (*models.Artist, error) {
	panic("not implemented")
}

func (m *MockServer) GetImageUrl(item models.Id, itemType models.ItemType) string {
	panic("not implemented")
}

func (m *MockServer) GetInfo() (*models.ServerInfo, error) {
	info := &models.ServerInfo{
		ServerType: "mock",
		Name:       "mock",
		Id:         "1234",
		Version:    "1.0",
	}
	return info, nil
}

func (m *MockServer) ConnectionOk() error {
	return nil
}

func (m *MockServer) GetConfig() config.Backend {
	return &MockConfig{}
}

func (m *MockServer) ReportProgress(state *interfaces.ApiPlaybackState) error {
	return nil
}

func (m *MockServer) Start() error {
	return nil
}

func (m *MockServer) Stop() error {
	return nil
}
