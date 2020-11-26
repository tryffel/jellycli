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
	"fmt"
	"io"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
)

func (jf *Jellyfin) Download(song *models.Song) (io.ReadCloser, interfaces.AudioFormat, error) {
	return jf.Stream(song)
}

func (jf *Jellyfin) Stream(song *models.Song) (rc io.ReadCloser, format interfaces.AudioFormat, err error) {
	format = interfaces.AudioFormatNil
	params := jf.defaultParams()
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
	jf.SessionId = util.RandomKey(20)
	ptr["PlaySessionId"] = jf.SessionId
	url := jf.host + "/Audio/" + song.Id.String() + "/universal"
	var stream *api.StreamBuffer
	stream, err = api.NewStreamDownload(url, map[string]string{"X-Emby-Token": jf.token}, *params, jf.client, song.Duration)
	rc = stream
	format, err = stream.AudioFormat()
	return
}
