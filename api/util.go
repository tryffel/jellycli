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

type infoResponse struct {
	ServerName string `json:"ServerName"`
	Version    string `json:"Version"`
}

// GetServerVersion returns name, version and possible error
func (a *Api) GetServerVersion() (string, string, error) {
	body, err := a.get("/System/Info/Public", nil)
	if err != nil {
		return "", "", fmt.Errorf("request failed: %v", err)
	}

	response := infoResponse{}
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return "", "", fmt.Errorf("response read failed: %v", err)
	}

	return response.ServerName, response.Version, nil
}
