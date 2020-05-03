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
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

func (a *Api) GetViews() ([]*models.View, error) {
	params := *a.defaultParams()

	url := fmt.Sprintf("/Users/%s/Views", a.userId)
	resp, err := a.get(url, &params)
	if err != nil {
		return nil, fmt.Errorf("get views: %v", err)
	}
	dto := views{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return nil, fmt.Errorf("parse views: %v", err)
	}

	views := make([]*models.View, len(dto.Views))
	for i, v := range dto.Views {
		views[i] = v.toView()
	}

	return views, nil
}

func (a *Api) GetLatestAlbums() ([]*models.Album, error) {
	params := *a.defaultParams()
	params["UserId"] = a.userId
	params.setParentId(a.musicView)

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items/Latest", a.userId), &params)
	if err != nil {
		return nil, fmt.Errorf("request latest albums: %v", err)
	}

	dto := []album{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return nil, fmt.Errorf("parse latest albums: %v", err)
	}

	albums := make([]*models.Album, len(dto))
	ids := make([]models.Id, len(dto))
	for i, v := range dto {
		albums[i] = v.toAlbum()
		ids[i] = albums[i].Id
	}
	a.cache.PutList("latest_music", ids)
	return albums, nil
}

func (a *Api) GetRecentlyPlayed(paging interfaces.Paging) ([]*models.Song, int, error) {
	params := *a.defaultParams()

	params.setIncludeTypes(mediaTypeSong)
	params.setSorting("DatePlayed", "Descending")
	params.enableRecursive()
	params["UserId"] = a.userId
	params.setParentId(a.musicView)

	if config.LimitRecentlyPlayed {
		paging = interfaces.Paging{
			CurrentPage: 0,
			PageSize:    config.LimitedRecentlyPlayedCount,
		}
	}
	params.setPaging(paging)

	resp, err := a.get(fmt.Sprintf("/Users/%s/Items", a.userId), &params)
	if err != nil {
		return nil, 0, fmt.Errorf("request latest albums: %v", err)
	}

	var dto songs
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return nil, 0, fmt.Errorf("parse latest albums: %v", err)
	}

	songs := make([]*models.Song, len(dto.Songs))
	for i, v := range dto.Songs {
		songs[i] = v.toSong()
	}

	totalSongs := dto.TotalSongs

	if config.LimitRecentlyPlayed {
		totalSongs = len(dto.Songs)
	}

	return songs, totalSongs, nil
}
