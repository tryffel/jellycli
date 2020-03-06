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
	"fmt"
	"io"
	"tryffel.net/go/jellycli/config"
)

func (a *Api) GetSongDirect(id string, codec string) (io.ReadCloser, error) {
	params := a.directplayParams()
	body, err := a.get("/Audio/"+id+"/stream.mp3", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}

	return body, nil
}

func (a *Api) directplayParams() *map[string]string {
	params := *(&map[string]string{})
	params["UserId"] = a.userId
	params["DeviceId"] = a.DeviceId
	params["MaxStreamingBitrate"] = "140000000"
	params["Container"] = "mp3"
	params["AudioSamplingRate"] = fmt.Sprint(config.AudioSamplingRate)

	// Every new request requires new playsession
	a.SessionId = randomKey(20)
	params["PlaySessionId"] = a.SessionId
	params["AudioCodec"] = "mp3"
	return &params
}
