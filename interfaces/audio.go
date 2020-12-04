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

package interfaces

type AudioFormat string

func (a AudioFormat) String() string {
	return string(a)
}

const (
	AudioFormatFlac AudioFormat = "flac"
	AudioFormatMp3  AudioFormat = "mp3"
	AudioFormatOgg  AudioFormat = "ogg"
	AudioFormatWav  AudioFormat = "wav"
	// empty format, for errors
	AudioFormatNil AudioFormat = ""
)

// SupportedAudioFormats
var SupportedAudioFormats = []AudioFormat{
	AudioFormatFlac,
	AudioFormatMp3,
	AudioFormatOgg,
	AudioFormatWav,
}
