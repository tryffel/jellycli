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
