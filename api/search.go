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
func (a *Api) Search(q string, limit int) (*SearchResult, error) {
	if limit == 0 {
		limit = 20
	}
	params := *a.defaultParams()
	delete(params, "UserId")
	delete(params, "DeviceId")
	params["SearchTerm"] = q
	params["Limit"] = fmt.Sprint(limit)
	params.setIncludeTypes(mediaTypeSong)

	body, err := a.get("/Search/Hints", &params)
	if err != nil {
		msg := getBodyMsg(body)
		return nil, fmt.Errorf("query failed: %v: %s", err, msg)
	}
	result := &SearchResult{}
	err = json.NewDecoder(body).Decode(result)
	if err != nil {
		err = fmt.Errorf("json parsing failed: %v", err)
	}

	if len(result.Items) > 0 {
		for i, _ := range result.Items {
			result.Items[i].Duration /= 10000000
		}
	}

	return result, err
}
