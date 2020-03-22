/*
 * Copyright 2020 Tero Vierimaa
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
