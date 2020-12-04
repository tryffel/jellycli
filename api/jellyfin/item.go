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

package jellyfin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
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

func (jf *Jellyfin) GetItem(id models.Id) (models.Item, error) {
	item, found := jf.cache.Get(id)
	if found && item != nil {
		return item, nil
	}
	params := jf.defaultParams()

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items/%s", jf.userId, id), params)
	if err != nil {
		return nil, fmt.Errorf("get item by id: %v", err)
	}

	b, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, fmt.Errorf("read http body: %v", err)
	}
	body := bytes.NewBuffer(b)

	dto := &map[string]interface{}{}
	err = json.NewDecoder(body).Decode(dto)
	if err != nil {
		return nil, fmt.Errorf("parse json response: %v", err)
	}

	itemT, err := getItemType(dto)
	if err != nil {
		return nil, fmt.Errorf("invalid item type: %v", err)
	}

	body = bytes.NewBuffer(b)

	item = nil
	switch itemT {
	case models.TypeSong:
		dto := song{}
		err = json.NewDecoder(body).Decode(&dto)
		if err != nil {
			return nil, fmt.Errorf("decode song: %v", err)
		}
		item = &(*dto.toSong())
	case models.TypeAlbum:
		dto := album{}
		err = json.NewDecoder(body).Decode(&dto)
		if err != nil {
			return nil, fmt.Errorf("decode album: %v", err)
		}
		item = &(*dto.toAlbum())
	case models.TypeArtist:
		dto := artist{}
		err = json.NewDecoder(body).Decode(&dto)
		if err != nil {
			return nil, fmt.Errorf("decode artist: %v", err)
		}
		item = &(*dto.toArtist())
	case models.TypePlaylist:
		dto := playlist{}
		err = json.NewDecoder(body).Decode(&dto)
		if err != nil {
			return nil, fmt.Errorf("decode playlist: %v", err)
		}
		item = &(*dto.toPlaylist())
	default:
		return nil, fmt.Errorf("unknown item type: %s", itemT)
	}

	jf.cache.Put(id, item, true)
	return item, nil
}

func (jf *Jellyfin) GetChildItems(id models.Id) ([]models.Item, error) {
	// get users/<uid>/items/<id>?parentid=<pid>
	return nil, nil
}

func (jf *Jellyfin) GetParentItem(id models.Id) (models.Item, error) {
	return nil, nil
}

func (jf *Jellyfin) GetArtist(id models.Id) (*models.Artist, error) {
	item, found := jf.cache.Get(id)
	// Return cached value if both artist and albums exist
	if found && item != nil {
		artist, ok := item.(*models.Artist)
		if !ok {
			jf.cache.Delete(id)
			logrus.Warningf("Found artist %s from cache with invalid type: %s", id, item.GetType())
		} else if artist.Albums != nil {
			if len(artist.Albums) == artist.AlbumCount {
				return artist, nil
			} else {
				jf.cache.Delete(id)
			}
		}
	}

	ar := &models.Artist{}

	params := jf.defaultParams()

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items/%s", jf.userId, id), params)
	if err != nil {
		return ar, fmt.Errorf("get artist: %v", err)
	}
	dto := artist{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return ar, fmt.Errorf("parse artist: %v", err)
	}

	ar = dto.toArtist()

	albums, err := jf.GetArtistAlbums(id)
	if err != nil {
		return ar, fmt.Errorf("get artist albums: %v", err)
	}

	ids := make([]models.Id, len(albums))
	items := make([]models.Item, len(albums))
	for i, v := range albums {
		ids[i] = v.Id
		items[i] = v
	}

	err = jf.cache.PutBatch(items, true)
	if err != nil {
		return ar, fmt.Errorf("store artist albums to cache: %v", err)
	}

	ar.Albums = ids
	jf.cache.Put(id, ar, true)

	return ar, nil
}

//GetArtistAlbums retrieves albums for given artist.
func (jf *Jellyfin) GetArtistAlbums(id models.Id) ([]*models.Album, error) {
	params := *jf.defaultParams()
	params.setIncludeTypes(mediaTypeAlbum)
	params.enableRecursive()
	//TODO: use also ContributingAlbumArtistIds
	params["AlbumArtistIds"] = id.String()
	params["Limit"] = defaultLimit
	params.setSorting("ProductionYear", "Ascending")

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), &params)
	if err != nil {
		return nil, fmt.Errorf("get artist albums: %v", err)
	}

	albums, _, err := jf.parseAlbums(resp)
	return albums, nil
}

func (jf *Jellyfin) GetAlbum(id models.Id) (*models.Album, error) {
	item, found := jf.cache.Get(id)
	// Return cached value if both artist and albums exist
	if found && item != nil {
		album, ok := item.(*models.Album)
		if !ok {
			jf.cache.Delete(id)
			logrus.Warningf("Found album %s from cache with invalid type: %s", id, item.GetType())
		} else if album.Songs != nil {
			if len(album.Songs) == album.SongCount {
				return album, nil
			} else {
				jf.cache.Delete(id)
			}
		}
	}

	al := &models.Album{}
	params := *jf.defaultParams()

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items/%s", jf.userId, id), &params)
	if err != nil {
		return al, fmt.Errorf("get album: %v", err)
	}
	dto := album{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return al, fmt.Errorf("parse album: %v", err)
	}

	al = dto.toAlbum()

	songs, err := jf.GetAlbumSongs(id)
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

	err = jf.cache.PutBatch(items, true)
	if err != nil {
		return al, fmt.Errorf("store artist albums to cache: %v", err)
	}
	al.SongCount = len(ids)
	al.Songs = ids
	jf.cache.Put(id, al, true)

	return al, nil
}

//GetAlbumSongs gets songs for given album.
func (jf *Jellyfin) GetAlbumSongs(album models.Id) ([]*models.Song, error) {
	params := *jf.defaultParams()
	params.enableRecursive()
	params.setParentId(album.String())
	params.setSorting("SortName", "Ascending")

	params["Limit"] = defaultLimit

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), &params)
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
		song := v.toSong()
		songs[i] = song
		jf.cache.Put(song.Id, song, true)
	}

	return songs, nil
}

func (jf *Jellyfin) GetFavoriteArtists() ([]*models.Artist, error) {
	params := *jf.defaultParams()
	params["IsFavorite"] = "true"

	resp, err := jf.get("/Artists", &params)
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

func (jf *Jellyfin) GetFavoriteAlbums(paging interfaces.Paging) ([]*models.Album, int, error) {
	params := jf.defaultParams()
	params.enableRecursive()
	params.setParentId(jf.musicView)
	params.setIncludeTypes(mediaTypeAlbum)
	params.setPaging(paging)
	ptr := params.ptr()
	ptr["Filters"] = "IsFavorite"

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return nil, 0, err
	}
	albums, count, err := jf.parseAlbums(resp)
	return albums, count, err
}

// GetPlaylists retrieves all playlists. Each playlists song count is known, but songs must be
// retrieved separately
func (jf *Jellyfin) GetPlaylists() ([]*models.Playlist, error) {
	params := *jf.defaultParams()
	params.setParentId(jf.musicView)
	params.setIncludeTypes(mediaTypePlaylist)
	params.enableRecursive()
	params["Fields"] = "ChildCount"

	data := make([]*models.Playlist, 0)

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), &params)
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
		pl := v.toPlaylist()
		data[i] = pl
		jf.cache.Put(pl.Id, pl, true)
	}

	return data, nil
}

// GetPlaylistSongs returns songs for playlist id
func (jf *Jellyfin) GetPlaylistSongs(playlist models.Id) ([]*models.Song, error) {
	params := *jf.defaultParams()
	params.setParentId(playlist.String())

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), &params)
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
func (jf *Jellyfin) GetSongs(page, pageSize int) ([]*models.Song, int, error) {
	params := *jf.defaultParams()
	params.setIncludeTypes(mediaTypeSong)
	params.enableRecursive()
	params.setSorting("Name", "Ascending")

	params["Limit"] = strconv.Itoa(pageSize)
	params["StartIndex"] = strconv.Itoa((page) * pageSize)

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), &params)
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

func (jf *Jellyfin) GetSongsById(ids []models.Id) ([]*models.Song, error) {
	params := *jf.defaultParams()
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

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), &params)
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

// getArtists return artists defined by paging and total number of artists
func (jf *Jellyfin) GetArtists(query *interfaces.QueryOpts) (artistList []*models.Artist, numRecords int, err error) {
	params := *jf.defaultParams()
	params.enableRecursive()
	params.setPaging(query.Paging)
	params.setSortingByType(models.TypeArtist, query.Sort)
	params.setFilter(models.TypeArtist, query.Filter)
	resp, err := jf.get("/Artists", &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return
	}
	return jf.parseArtists(resp)
}

// getArtists return artists defined by paging and total number of artists
func (jf *Jellyfin) getArtists(paging interfaces.Paging) (artistList []*models.Artist, numRecords int, err error) {
	params := *jf.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setPaging(paging)
	resp, err := jf.get("/Artists", &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return
	}

	return jf.parseArtists(resp)
}

func (jf *Jellyfin) GetAlbumArtists(paging interfaces.Paging) (artistList []*models.Artist, numRecords int, err error) {
	params := *jf.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setPaging(paging)
	resp, err := jf.get("/Artists/AlbumArtists", &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return
	}
	return jf.parseArtists(resp)
}

func (jf *Jellyfin) parseArtists(resp io.Reader) (artistList []*models.Artist, numRecords int, err error) {
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

func (jf *Jellyfin) parseAlbums(resp io.Reader) (albumList []*models.Album, numRecords int, err error) {
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
func (jf *Jellyfin) GetAlbums(opts *interfaces.QueryOpts) (albumList []*models.Album, numRecords int, err error) {
	params := *jf.defaultParams()
	params.enableRecursive()
	params.setPaging(opts.Paging)
	params.setSortingByType(models.TypeAlbum, opts.Sort)
	params.setFilter(models.TypeAlbum, opts.Filter)
	params.setIncludeTypes(mediaTypeAlbum)
	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return
	}

	return jf.parseAlbums(resp)
}

func (jf *Jellyfin) GetSimilarArtists(artist models.Id) ([]*models.Artist, error) {
	params := *jf.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setLimit(50)
	resp, err := jf.get(fmt.Sprintf("/Items/%s/Similar", artist.String()), &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return []*models.Artist{}, err
	}

	artists, _, err := jf.parseArtists(resp)
	return artists, err
}

func (jf *Jellyfin) GetSimilarAlbums(album models.Id) ([]*models.Album, error) {
	params := *jf.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setLimit(50)
	resp, err := jf.get(fmt.Sprintf("/Items/%s/Similar", album.String()), &params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return []*models.Album{}, err
	}

	albums, _, err := jf.parseAlbums(resp)
	return albums, err

}

func (jf *Jellyfin) GetGenres(paging interfaces.Paging) ([]*models.IdName, int, error) {
	params := jf.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setPaging(paging)
	params.setParentId(jf.musicView)

	resp, err := jf.get("/Genres", params)
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

func (jf *Jellyfin) GetGenreAlbums(genre models.IdName) ([]*models.Album, error) {
	params := jf.defaultParams()
	params.enableRecursive()
	params.setSorting("SortName", "Ascending")
	params.setParentId(jf.musicView)
	(*params)["GenreIds"] = genre.Id.String()
	params.setIncludeTypes(mediaTypeAlbum)

	resp, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userId), params)
	if resp != nil {
		defer resp.Close()
	}

	if err != nil {
		return []*models.Album{}, err
	}

	albums, _, err := jf.parseAlbums(resp)
	return albums, err
}

func (jf *Jellyfin) GetAlbumArtist(album *models.Album) (*models.Artist, error) {
	artist := jf.cache.GetArtist(album.Id)
	if artist == nil {
		artist, err := jf.GetArtist(album.Artist)
		if err != nil {
			return nil, fmt.Errorf("get artist: %v", err)
		}
		jf.cache.Put(artist.Id, artist, true)
		return artist, nil
	}
	return artist, nil
}
