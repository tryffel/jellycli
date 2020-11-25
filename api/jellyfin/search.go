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
func (a *Api) Search(query string, itemType models.ItemType, limit int) ([]models.Item, error) {
	if limit == 0 {
		limit = 40
	}
	params := *a.defaultParams()
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
		url = fmt.Sprintf("/Users/%s/Items", a.userId)
	case models.TypeSong:
		params.setIncludeTypes(mediaTypeSong)
		url = fmt.Sprintf("/Users/%s/Items", a.userId)
	case models.TypePlaylist:
		params.setIncludeTypes(mediaTypePlaylist)
		url = fmt.Sprintf("/Users/%s/Items", a.userId)
	case models.TypeGenre:
		return nil, errors.New("genres not supported")
	}

	body, err := a.get(url, &params)
	if err != nil {
		msg := getBodyMsg(body)
		return nil, fmt.Errorf("query failed: %v: %s", err, msg)
	}

	return searchDtoToItems(body, toItemType(itemType))
}
