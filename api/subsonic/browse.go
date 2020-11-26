/*
 * Copyright 2020 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package subsonic

import (
	"errors"
	"strconv"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

func (s *Subsonic) getFavorites() error {
	if len(s.favoriteAlbums) == 0 || len(s.favoriteArtists) == 0 {
		resp, err := s.get("/getStarred2", nil)
		if err != nil {
			return err
		}

		s.favoriteAlbums = make([]*models.Album, len(resp.Favorites.Albums))
		for i, v := range resp.Favorites.Albums {
			s.favoriteAlbums[i] = v.toAlbum()
		}
		s.favoriteArtists = make([]*models.Artist, len(resp.Favorites.Artists))
		for i, v := range resp.Favorites.Artists {
			s.favoriteArtists[i] = v.toArtist()
		}
	}
	return nil
}

func (s *Subsonic) GetArtists(paging interfaces.Paging) (artists []*models.Artist, n int, err error) {
	var resp *response
	resp, err = s.get("/getArtists", nil)
	if err != nil {
		return nil, 0, err
	}

	i := 0

	for _, indexV := range *resp.Artists.Indexes {
		artists = append(artists, make([]*models.Artist, len(*indexV.Artists))...)
		for _, v := range *indexV.Artists {
			artists[i] = v.toArtist()
			i += 1
		}
	}

	return artists, len(artists), nil
}

func (s *Subsonic) GetAlbumArtists(paging interfaces.Paging) ([]*models.Artist, int, error) {
	return s.GetArtists(paging)
}

func (s *Subsonic) GetAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {

	params := &params{}
	(*params)["type"] = "alphabeticalByName"
	(*params)["offset"] = strconv.Itoa(paging.Offset())
	(*params)["size"] = strconv.Itoa(paging.PageSize)

	resp, err := s.get("/getAlbumList2", params)
	if err != nil {
		return nil, 0, err
	}

	albums := make([]*models.Album, len(resp.AlbumList.Albums))

	for i, v := range resp.AlbumList.Albums {
		albums[i] = v.toAlbum()
	}
	return albums, len(albums), nil
}

func (s *Subsonic) GetArtistAlbums(artist models.Id) (albums []*models.Album, err error) {
	params := &params{}
	params.setId(artist.String())
	resp, err := s.get("/getArtist", params)
	if err != nil {
		return nil, err
	}

	albums = make([]*models.Album, len(resp.Artist.Albums))
	for i, v := range resp.Artist.Albums {
		albums[i] = v.toAlbum()
	}
	return albums, nil

}

func (s *Subsonic) GetAlbumSongs(album models.Id) ([]*models.Song, error) {

	params := &params{}
	params.setId(album.String())

	resp, err := s.get("/getAlbum", params)
	if err != nil {
		return nil, err
	}

	songs := make([]*models.Song, len(resp.Albums.Songs))

	for i, v := range resp.Albums.Songs {
		song := v.toSong()
		songs[i] = song
	}
	return songs, nil

}

func (s *Subsonic) GetPlaylists() ([]*models.Playlist, error) {
	panic("implement me")
}

func (s *Subsonic) GetPlaylistSongs(playlist models.Id) ([]*models.Song, error) {
	panic("implement me")
}

func (s *Subsonic) GetFavoriteArtists() ([]*models.Artist, error) {
	err := s.getFavorites()
	return s.favoriteArtists, err
}

func (s *Subsonic) GetFavoriteAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	err := s.getFavorites()
	return s.favoriteAlbums, len(s.favoriteAlbums), err
}

func (s *Subsonic) GetSimilarArtists(artist models.Id) ([]*models.Artist, error) {
	panic("implement me")
}

func (s *Subsonic) GetSimilarAlbums(album models.Id) ([]*models.Album, error) {
	panic("implement me")
}

func (s *Subsonic) GetLatestAlbums() ([]*models.Album, error) {
	panic("implement me")
}

func (s *Subsonic) GetRecentlyPlayed(paging interfaces.Paging) ([]*models.Song, int, error) {
	panic("implement me")
}

func (s *Subsonic) GetServerInfo() api.ServerInfo {
	panic("implement me")
}

func (s *Subsonic) GetSongs(page, pageSize int) ([]*models.Song, int, error) {
	panic("implement me")
}

func (s *Subsonic) GetGenres(paging interfaces.Paging) ([]*models.IdName, int, error) {
	panic("implement me")
}

func (s *Subsonic) GetGenreAlbums(genre models.IdName) ([]*models.Album, error) {
	panic("implement me")
}

func (s *Subsonic) GetAlbumArtist(album *models.Album) (*models.Artist, error) {
	params := &params{}
	params.setId(album.Artist.String())
	resp, err := s.get("/getArtist", params)
	if err != nil {
		return nil, err
	}

	if resp.Artist == nil {
		return nil, nil
	}

	artist := resp.Artist.toArtist()
	return artist, nil
}

func (s *Subsonic) GetSongArtistAlbum(song *models.Song) (*models.Album, *models.Artist, error) {
	if song.AlbumArtist == "" {
		return nil, nil, errors.New("no album artist defined")
	}

	params := &params{}
	params.setId(song.AlbumArtist.String())
	resp, err := s.get("/getArtist", params)
	if err != nil {
		return nil, nil, err
	}

	if resp.Artist == nil {
		return nil, nil, nil
	}

	artist := resp.Artist.toArtist()
	var album *models.Album

	for _, v := range resp.Artist.Albums {
		if v.Id == song.Album.String() {
			album = v.toAlbum()
			break
		}
	}
	return album, artist, nil
}

func (s *Subsonic) GetInstantMix(item models.Item) ([]*models.Song, error) {
	panic("implement me")
}

func (s *Subsonic) GetLink(item models.Item) string {
	panic("implement me")
}

func (s *Subsonic) Search(query string, itemType models.ItemType, maxResults int) ([]models.Item, error) {
	params := &params{}
	(*params)["query"] = query
	(*params)["artistCount"] = "0"
	(*params)["albumCount"] = "0"
	(*params)["songCount"] = "0"

	limit := strconv.Itoa(maxResults)

	switch itemType {
	case models.TypeArtist:
		(*params)["artistCount"] = limit
	case models.TypeAlbum:
		(*params)["albumCount"] = limit
	case models.TypeSong:
		(*params)["songCount"] = limit
	}

	resp, err := s.get("/search3", params)
	if err != nil {
		return nil, err
	}

	if resp.Search == nil {
		return nil, nil
	}

	var items []models.Item

	if itemType == models.TypeArtist {
		items = make([]models.Item, len(resp.Search.Artists))
		for i, v := range resp.Search.Artists {
			items[i] = v.toArtist()
		}
	} else if itemType == models.TypeAlbum {
		items = make([]models.Item, len(resp.Search.Albums))
		for i, v := range resp.Search.Albums {
			items[i] = v.toAlbum()
		}
	} else if itemType == models.TypeSong {
		items = make([]models.Item, len(resp.Search.Songs))
		for i, v := range resp.Search.Songs {
			items[i] = v.toSong()
		}
	}

	return items, nil
}
