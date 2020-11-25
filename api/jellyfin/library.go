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

func (a *Api) GetUserViews() {
	body, err := a.get("/Users/"+a.userId+"/Views", nil)
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
