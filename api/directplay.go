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
	"net/http"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
)

func (a *Api) GetSongUniversal(id string) (rc io.ReadCloser, format interfaces.AudioFormat, err error) {
	format = interfaces.AudioFormatNil
	params := a.defaultParams()
	ptr := params.ptr()
	ptr["MaxStreamingBitrate"] = "140000000"
	ptr["AudioSamplingRate"] = fmt.Sprint(config.AudioSamplingRate)
	formats := ""
	for i, v := range interfaces.SupportedAudioFormats {
		if i > 0 {
			formats += ","
		}
		formats += v.String()
	}
	ptr["Container"] = formats
	// Every new request requires new playsession
	a.SessionId = randomKey(20)
	ptr["PlaySessionId"] = a.SessionId
	ptr["AudioCodec"] = "mp3"
	resp, err := a.makeRequest(http.MethodGet, "/Audio/"+id+"/universal", nil, params)
	if err != nil {
		err = fmt.Errorf("download file: %v", err)
		return
	}

	format, err = mimeToAudioFormat(resp.Header.Get("Content-Type"))
	rc = resp.Body
	return
}

func mimeToAudioFormat(mimeType string) (format interfaces.AudioFormat, err error) {
	format = interfaces.AudioFormatNil
	switch mimeType {
	case "audio/mpeg":
		format = interfaces.AudioFormatMp3
	case "audio/flac":
		format = interfaces.AudioFormatFlac
	case "audio/ogg":
		format = interfaces.AudioFormatOgg
	case "audio/wav":
		format = interfaces.AudioFormatWav

	default:
		err = fmt.Errorf("unidentified audio format: %s", mimeType)
	}
	return
}
