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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"tryffel.net/go/jellycli/models"
)

type SearchHint struct {
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Year        int    `json:"ProductionYear"`
	Type        string `json:"Type"`
	Duration    int    `json:"RunTimeTicks"`
	Album       string `json:"Album"`
	AlbumId     string `json:"AlbumId"`
	AlbumArtist string `json:"AlbumArtist"`
}

type SearchResult struct {
	Items []SearchHint `json:"SearchHints"`
}

func searchDtoToItems(rc io.ReadCloser, target mediaItemType) ([]models.Item, error) {

	var result itemMapper

	switch target {
	case mediaTypeSong:
		result = &songs{}
	case mediaTypeAlbum:
		result = &albums{}
	case mediaTypeArtist:
		result = &artists{}
	case mediaTypePlaylist:
		result = &playlists{}
	default:
		return nil, fmt.Errorf("unknown item type: %s", target)
	}

	err := json.NewDecoder(rc).Decode(result)
	if err != nil {
		return nil, fmt.Errorf("decode item %s: %v", target, err)
	}

	return result.Items(), nil
}

//Search searches audio items
func (jf *Jellyfin) Search(query string, itemType models.ItemType, limit int) ([]models.Item, error) {
	if limit == 0 {
		limit = 40
	}
	params := *jf.defaultParams()
	params.enableRecursive()
	params["SearchTerm"] = query
	params["Limit"] = fmt.Sprint(limit)
	params["IncludePeople"] = "false"
	params["IncludeMedia"] = "true"
	var url string

	switch itemType {
	case models.TypeArtist:
		params["IncludeArtists"] = "true"
		params["IncludeMedia"] = "false"
		url = "/Artists"
	case models.TypeAlbum:
		params.setIncludeTypes(mediaTypeAlbum)
		url = fmt.Sprintf("/Users/%s/Items", jf.userId)
	case models.TypeSong:
		params.setIncludeTypes(mediaTypeSong)
		url = fmt.Sprintf("/Users/%s/Items", jf.userId)
	case models.TypePlaylist:
		params.setIncludeTypes(mediaTypePlaylist)
		url = fmt.Sprintf("/Users/%s/Items", jf.userId)
	case models.TypeGenre:
		return nil, errors.New("genres not supported")
	}

	body, err := jf.get(url, &params)
	if err != nil {
		msg := getBodyMsg(body)
		return nil, fmt.Errorf("query failed: %v: %s", err, msg)
	}

	return searchDtoToItems(body, toItemType(itemType))
}
