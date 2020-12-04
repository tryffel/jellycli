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
	"fmt"
)

type MediaViewResponse struct {
	Libraries []MediaLibrary `json:"Items"`
}

type MediaLibrary struct {
	Name     string `json:"Name"`
	ServerId string `json:"ServerId"`
	Id       string `json:"Id"`
}

func (jf *Jellyfin) GetUserViews() {
	body, err := jf.get("/Users/"+jf.userId+"/Views", nil)
	if err != nil {
		println(fmt.Errorf("failed to get views: %v", err))
	}

	resp := &MediaViewResponse{}
	err = json.NewDecoder(body).Decode(&resp)
	if err != nil {
		fmt.Printf("Invalid server response: %v", err)
	} else {
		fmt.Println(resp)
	}
}
