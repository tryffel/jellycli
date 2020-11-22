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
	"errors"
	"fmt"
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

//Search searches audio items
func (a *Api) Search(query string, itemType models.ItemType, limit int) ([]models.Item, error) {
	if limit == 0 {
		limit = 40
	}
	params := *a.defaultParams()
	params.enableRecursive()
	delete(params, "UserId")
	delete(params, "DeviceId")
	params["SearchTerm"] = query
	params["Limit"] = fmt.Sprint(limit)
	params["IncludePeople"] = "false"
	params["IncludeMedia"] = "true"
	var url string

	switch itemType {
	case models.TypeArtist:
		params.setIncludeTypes(mediaTypeArtist)
		url = "/Artists"
	case models.TypeAlbum:
		params.setIncludeTypes(mediaTypeAlbum)
		url = "/Albums"
		url = fmt.Sprintf("/Users/%s/Items", a.userId)
	case models.TypeSong:
		params.setIncludeTypes(mediaTypeSong)
		url = fmt.Sprintf("/Users/%s/Items", a.userId)
	case models.TypePlaylist:
		params.setIncludeTypes(mediaTypePlaylist)
		url = fmt.Sprintf("/Users/%s/Items", a.userId)
		url = "/Playlists"
	case models.TypeGenre:
		return nil, errors.New("genres not supported")
	}

	type Result struct {
		Items            []album
		TotalRecordCount int
	}

	result := &Result{}

	body, err := a.get(url, &params)
	if err != nil {
		msg := getBodyMsg(body)
		return nil, fmt.Errorf("query failed: %v: %s", err, msg)
	}

	err = json.NewDecoder(body).Decode(result)
	if err != nil {
		return []models.Item{}, err
	}

	items := make([]models.Item, len(result.Items))

	for i, v := range result.Items {
		items[i] = v.ModelType()
	}
	return items, nil
}
