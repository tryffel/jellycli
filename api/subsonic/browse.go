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

package subsonic

import (
	"errors"
	"strconv"
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

func (s *Subsonic) getAlbums(params *params) ([]*models.Album, error) {
	resp, err := s.get("/getAlbumList2", params)
	if err != nil {
		return nil, err
	}

	albums := make([]*models.Album, len(resp.AlbumList.Albums))

	for i, v := range resp.AlbumList.Albums {
		albums[i] = v.toAlbum()
	}
	return albums, nil
}

func (s *Subsonic) GetAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	params := &params{}
	(*params)["type"] = "alphabeticalByName"
	(*params)["offset"] = strconv.Itoa(paging.Offset())
	(*params)["size"] = strconv.Itoa(paging.PageSize)
	albums, err := s.getAlbums(params)
	return albums, len(albums), err
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
	resp, err := s.get("/getPlaylists", nil)
	if err != nil {
		return nil, err
	}
	playlists := make([]*models.Playlist, len(resp.Playlists.Playlists))
	for i, v := range resp.Playlists.Playlists {
		playlists[i] = v.toPlaylist()
	}
	return playlists, nil
}

func (s *Subsonic) GetPlaylistSongs(playlist models.Id) ([]*models.Song, error) {
	params := &params{}
	params.setId(playlist.String())
	resp, err := s.get("/getPlaylist", params)
	if err != nil {
		return nil, err
	}
	songs := make([]*models.Song, len(resp.Playlist.Songs))
	for i, v := range resp.Playlist.Songs {
		songs[i] = v.toSong()
	}
	return songs, nil
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

	return nil, errors.New("not implemented")
}

func (s *Subsonic) GetSimilarAlbums(album models.Id) ([]*models.Album, error) {
	return nil, errors.New("not implemented")
}

func (s *Subsonic) GetLatestAlbums() ([]*models.Album, error) {
	params := &params{}
	(*params)["type"] = "newest"
	return s.getAlbums(params)
}

func (s *Subsonic) GetRecentlyPlayed(paging interfaces.Paging) ([]*models.Song, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (s *Subsonic) GetSongs(page, pageSize int) ([]*models.Song, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (s *Subsonic) GetGenres(paging interfaces.Paging) ([]*models.IdName, int, error) {
	resp, err := s.get("/getGenres", nil)
	if err != nil {
		return nil, 0, err
	}
	genres := make([]*models.IdName, len(resp.Genres.Genres))
	for i, v := range resp.Genres.Genres {
		genres[i] = v.toGenre()
	}
	return genres, len(genres), nil
}

func (s *Subsonic) GetGenreAlbums(genre models.IdName) ([]*models.Album, error) {
	params := &params{}
	(*params)["type"] = "byGenre"
	(*params)["genre"] = genre.Name
	albums, err := s.getAlbums(params)
	return albums, err
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

func (s *Subsonic) GetInstantMix(item models.Item) ([]*models.Song, error) {
	return nil, errors.New("not implemented")
}

func (s *Subsonic) GetLink(item models.Item) string {
	return ""
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

func (s *Subsonic) GetAlbum(id models.Id) (*models.Album, error) {
	params := &params{}
	params.setId(id.String())

	resp, err := s.get("/getAlbum", params)
	if err != nil {
		return nil, err
	}

	album := resp.Albums.toAlbum()
	return album, nil
}

func (s *Subsonic) GetArtist(id models.Id) (*models.Artist, error) {
	params := &params{}
	params.setId(id.String())

	resp, err := s.get("/getArtist", params)
	if err != nil {
		return nil, err
	}

	artist := resp.Artist.toArtist()
	return artist, nil
}

func (s *Subsonic) ImageUrl(item models.Id, itemType models.ItemType) string {
	return ""
}
