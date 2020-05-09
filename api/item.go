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
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"strconv"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

const (
	defaultLimit = "100"
)

func getItemType(dto *map[string]interface{}) (models.ItemType, error) {
	field := (*dto)["Type"]
	text, ok := field.(string)
	if !ok {
		return "", fmt.Errorf("invalid type: %v", text)
	}
	switch text {
	case mediaTypeArtist.String():
		return models.TypeArtist, nil
	case mediaTypeAlbum.String():
		return models.TypeAlbum, nil
	case mediaTypeSong.String():
		return models.TypeSong, nil
	default:
		return "", fmt.Errorf("unknown type: %s", text)
	}
}

func (a *Api) GetItem(id models.Id) (models.Item, error) {
	item, found := a.cache.Get(id)
	if found && item != nil {
		return item, nil
	}
	params := a.defaultParams()

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items/%s", a.userId, id), params)
	if err != nil {
		return nil, fmt.Errorf("get item by id: %v", err)
	}
	dto := &map[string]interface{}{}
	err = json.NewDecoder(resp).Decode(dto)
	if err != nil {
		return nil, fmt.Errorf("parse json response: %v", err)
	}

	itemT, err := getItemType(dto)
	if err != nil {
		return nil, fmt.Errorf("invalid item type: %v", err)
	}
	//decoder := json.NewDecoder(resp)
	//var item models.Item
	switch itemT {
	case models.TypeAlbum:

	case models.TypeArtist:
	}
	return nil, nil
}

func (a *Api) GetChildItems(id models.Id) ([]models.Item, error) {
	// get users/<uid>/items/<id>?parentid=<pid>
	return nil, nil
}

func (a *Api) GetParentItem(id models.Id) (models.Item, error) {
	return nil, nil
}

func (a *Api) GetArtist(id models.Id) (models.Artist, error) {
	item, found := a.cache.Get(id)
	// Return cached value if both artist and albums exist
	if found && item != nil {
		artist, ok := item.(*models.Artist)
		if !ok {
			a.cache.Delete(id)
			logrus.Warningf("Found artist %s from cache with invalid type: %s", id, item.GetType())
		} else if artist.Albums != nil {
			if len(artist.Albums) == artist.AlbumCount {
				return *artist, nil
			} else {
				a.cache.Delete(id)
			}
		}
	}

	ar := models.Artist{}

	params := a.defaultParams()

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items/%s", a.userId, id), params)
	if err != nil {
		return ar, fmt.Errorf("get artist: %v", err)
	}
	dto := artist{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return ar, fmt.Errorf("parse artist: %v", err)
	}

	ar = *dto.toArtist()

	albums, err := a.GetArtistAlbums(id)
	if err != nil {
		return ar, fmt.Errorf("get artist albums: %v", err)
	}

	ids := make([]models.Id, len(albums))
	items := make([]models.Item, len(albums))
	for i, v := range albums {
		ids[i] = v.Id
		items[i] = v
	}

	err = a.cache.PutBatch(items, true)
	if err != nil {
		return ar, fmt.Errorf("store artist albums to cache: %v", err)
	}

	ar.Albums = ids
	a.cache.Put(id, &ar, true)

	return ar, nil
}

//GetArtistAlbums retrieves albums for given artist.
func (a *Api) GetArtistAlbums(id models.Id) ([]*models.Album, error) {
	params := *a.defaultParams()
	params.setIncludeTypes(mediaTypeAlbum)
	params.enableRecursive()
	//TODO: use also ContributingAlbumArtistIds
	params["AlbumArtistIds"] = id.String()
	params["Limit"] = defaultLimit
	params.setSorting("ProductionYear", "Ascending")

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), &params)
	if err != nil {
		return nil, fmt.Errorf("get artist albums: %v", err)
	}

	albums, _, err := a.parseAlbums(resp)
	return albums, nil
}

func (a *Api) GetAlbum(id models.Id) (models.Album, error) {
	item, found := a.cache.Get(id)
	// Return cached value if both artist and albums exist
	if found && item != nil {
		album, ok := item.(*models.Album)
		if !ok {
			a.cache.Delete(id)
			logrus.Warningf("Found album %s from cache with invalid type: %s", id, item.GetType())
		} else if album.Songs != nil {
			if len(album.Songs) == album.SongCount {
				return *album, nil
			} else {
				a.cache.Delete(id)
			}
		}
	}

	al := models.Album{}
	params := *a.defaultParams()

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items/%s", a.userId, id), &params)
	if err != nil {
		return al, fmt.Errorf("get album: %v", err)
	}
	dto := album{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return al, fmt.Errorf("parse album: %v", err)
	}

	al = *dto.toAlbum()

	songs, err := a.GetAlbumSongs(id)
	if err != nil {
		return al, fmt.Errorf("get albums songs: %v", err)
	}

	for _, v := range songs {
		v.AlbumArtist = al.Artist
	}

	ids := make([]models.Id, len(songs))
	items := make([]models.Item, len(songs))
	for i, v := range songs {
		ids[i] = v.Id
		items[i] = v
	}

	err = a.cache.PutBatch(items, true)
	if err != nil {
		return al, fmt.Errorf("store artist albums to cache: %v", err)
	}
	al.SongCount = len(ids)
	al.Songs = ids
	a.cache.Put(id, &al, true)

	return al, nil
}

//GetAlbumSongs gets songs for given album.
func (a *Api) GetAlbumSongs(album models.Id) ([]*models.Song, error) {
	params := *a.defaultParams()
	params.enableRecursive()
	params.setParentId(album.String())
	params.setSorting("SortName", "Ascending")

	params["Limit"] = defaultLimit

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), &params)
	if err != nil {
		return nil, fmt.Errorf("get album Songs; %v", err)
	}

	dto := songs{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Songs: %v", err)
	}

	songs := make([]*models.Song, len(dto.Songs))
	for i, v := range dto.Songs {
		songs[i] = v.toSong()
	}

	return songs, nil
}

func (a *Api) GetFavoriteArtists() ([]*models.Artist, error) {
	params := *a.defaultParams()
	params["IsFavorite"] = "true"

	resp, err := a.get("/Artists", &params)
	if err != nil {
		return nil, fmt.Errorf("get favorite artists: %v", err)
	}

	dto := artists{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return nil, fmt.Errorf("parse artists: %v", err)
	}

	artists := make([]*models.Artist, len(dto.Artists))

	// FavoriteArtists doesn't return any album info
	for i, v := range dto.Artists {
		if v.TotalAlbums == 0 {
			v.TotalAlbums = -1
		}
		artists[i] = v.toArtist()
	}
	return artists, nil
}

func (a *Api) GetFavoriteAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	params := a.defaultParams()
	params.enableRecursive()
	params.setParentId(a.musicView)
	params.setIncludeTypes(mediaTypeAlbum)
	params.setPaging(paging)
	ptr := params.ptr()
	ptr["Filters"] = "IsFavorite"

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return nil, 0, err
	}
	albums, count, err := a.parseAlbums(resp)
	return albums, count, err
}

// GetPlaylists retrieves all playlists. Each playlists song count is known, but songs must be
// retrieved separately
func (a *Api) GetPlaylists() ([]*models.Playlist, error) {
	params := *a.defaultParams()
	params.setParentId(a.musicView)
	params.setIncludeTypes(mediaTypePlaylist)
	params.enableRecursive()
	params["Fields"] = "ChildCount"

	data := make([]*models.Playlist, 0)

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), &params)
	if resp != nil {
		defer resp.Close()
	}

	dto := playlists{}
	err = json.NewDecoder(resp).Decode(&dto)

	if err != nil {
		return data, fmt.Errorf("parse playlists: %v", err)
	}

	data = make([]*models.Playlist, len(dto.Playlists))
	for i, v := range dto.Playlists {
		logInvalidType(&v, "get playlists")
		data[i] = v.toPlaylist()
	}

	return data, nil
}

// GetPlaylistSongs returns songs for playlist id
func (a *Api) GetPlaylistSongs(playlist models.Id) ([]*models.Song, error) {
	params := *a.defaultParams()
	params.setParentId(playlist.String())

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return []*models.Song{}, err
	}

	dto := songs{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return []*models.Song{}, fmt.Errorf("decode json: %v", err)
	}

	songs := make([]*models.Song, len(dto.Songs))
	for i, v := range dto.Songs {
		logInvalidType(&v, "get playlist songs")
		songs[i] = v.toSong()
	}

	return songs, nil
}

// GetSongs returns songs by paging, and returns total number of songs
func (a *Api) GetSongs(page, pageSize int) ([]*models.Song, int, error) {
	params := *a.defaultParams()
	params.setIncludeTypes(mediaTypeSong)
	params.enableRecursive()
	params.setSorting("Name", "Ascending")

	params["Limit"] = strconv.Itoa(pageSize)
	params["StartIndex"] = strconv.Itoa((page) * pageSize)

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), &params)
	if resp != nil {
		defer resp.Close()
	}

	if err != nil {
		return []*models.Song{}, 0, err
	}

	dto := songs{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return []*models.Song{}, 0, fmt.Errorf("decode json: %v", err)
	}

	songs := make([]*models.Song, len(dto.Songs))

	for i, v := range dto.Songs {
		logInvalidType(&v, "get songs")
		songs[i] = v.toSong()
		songs[i].Index = pageSize*pageSize + i + 1
	}

	return songs, dto.TotalSongs, nil
}

func (a *Api) GetSongsById(ids []models.Id) ([]*models.Song, error) {
	params := *a.defaultParams()
	params.setIncludeTypes(mediaTypeSong)
	params.enableRecursive()

	if len(ids) == 0 {
		return []*models.Song{}, fmt.Errorf("ids cannot be empty")
	}

	idList := ""
	for i, v := range ids {
		if i > 0 {
			idList += ","
		}
		idList += v.String()
	}

	params["Ids"] = idList

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), &params)
	if resp != nil {
		defer resp.Close()
	}

	if err != nil {
		return []*models.Song{}, err
	}

	dto := songs{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return []*models.Song{}, fmt.Errorf("decode json: %v", err)
	}

	songs := make([]*models.Song, len(dto.Songs))

	for i, v := range dto.Songs {
		logInvalidType(&v, "get songs")
		songs[i] = v.toSong()
		songs[i].Index = i + 1
	}

	return songs, nil
}

// GetArtists return artists defined by paging and total number of artists
func (a *Api) GetArtists(paging interfaces.Paging) (artistList []*models.Artist, numRecords int, err error) {
	params := *a.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setPaging(paging)
	resp, err := a.get("/Artists", &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return
	}

	return a.parseArtists(resp)
}

func (a *Api) GetAlbumArtists(paging interfaces.Paging) (artistList []*models.Artist, numRecords int, err error) {
	params := *a.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setPaging(paging)
	resp, err := a.get("/Artists/AlbumArtists", &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return
	}
	return a.parseArtists(resp)
}

func (a *Api) parseArtists(resp io.Reader) (artistList []*models.Artist, numRecords int, err error) {
	dto := &artists{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		err = fmt.Errorf("decode json: %v", err)
		return
	}

	artistList = make([]*models.Artist, len(dto.Artists))
	for i, v := range dto.Artists {
		logInvalidType(&v, "get artists")
		artistList[i] = v.toArtist()
	}

	numRecords = dto.TotalArtists
	return
}

func (a *Api) parseAlbums(resp io.Reader) (albumList []*models.Album, numRecords int, err error) {
	dto := albums{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		err = fmt.Errorf("decode json: %v", err)
		return
	}

	numRecords = dto.TotalAlbums
	albumList = make([]*models.Album, len(dto.Albums))
	for i, v := range dto.Albums {
		logInvalidType(&v, "get albums")
		albumList[i] = v.toAlbum()
	}
	return
}

// GetAlbums returns albums with given paging. It also returns number of all albums
func (a *Api) GetAlbums(paging interfaces.Paging) (albumList []*models.Album, numRecords int, err error) {
	params := *a.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setPaging(paging)
	params.setIncludeTypes(mediaTypeAlbum)
	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return
	}

	return a.parseAlbums(resp)
}

func (a *Api) GetSimilarArtists(artist models.Id) ([]*models.Artist, error) {
	params := *a.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setLimit(50)
	resp, err := a.get(fmt.Sprintf("/Items/%s/Similar", artist.String()), &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return []*models.Artist{}, err
	}

	artists, _, err := a.parseArtists(resp)
	return artists, err
}

func (a *Api) GetSimilarAlbums(album models.Id) ([]*models.Album, error) {
	params := *a.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setLimit(50)
	resp, err := a.get(fmt.Sprintf("/Items/%s/Similar", album.String()), &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return []*models.Album{}, err
	}

	albums, _, err := a.parseAlbums(resp)
	return albums, err

}

func (a *Api) GetGenres(paging interfaces.Paging) ([]*models.IdName, int, error) {
	params := a.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setPaging(paging)
	params.setParentId(a.musicView)

	resp, err := a.get("/Genres", params)
	if resp != nil {
		defer resp.Close()
	}

	if err != nil {
		return []*models.IdName{}, 0, err
	}

	body := struct {
		Items []nameId
		Count int `json:"TotalRecordCount"`
	}{}

	ids := make([]*models.IdName, 0)
	err = json.NewDecoder(resp).Decode(&body)
	if err != nil {
		return ids, 0, fmt.Errorf("decode json: %v", err)
	}

	ids = make([]*models.IdName, len(body.Items))
	for i, v := range body.Items {
		ids[i] = &models.IdName{Id: models.Id(v.Id), Name: v.Name}
	}

	return ids, body.Count, nil
}

func (a *Api) GetGenreAlbums(genre models.IdName) ([]*models.Album, error) {
	params := a.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setParentId(a.musicView)
	(*params)["GenreIds"] = genre.Id.String()
	params.setIncludeTypes(mediaTypeAlbum)

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), params)
	if resp != nil {
		defer resp.Close()
	}

	if err != nil {
		return []*models.Album{}, err
	}

	albums, _, err := a.parseAlbums(resp)
	return albums, err
}
