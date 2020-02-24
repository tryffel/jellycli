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

package mpris

import (
	"fmt"
	"tryffel.net/go/jellycli/controller"

	"github.com/godbus/dbus"
)

// TrackIDFormat is the formatter string for a track ID.
const TrackIDFormat = "/org/mpd/Tracks/%s"

// This file implements a struct that satisfies the `org.mpris.MediaPlayer2.TrackList` interface.

// TrackList is a DBus object satisfying the `org.mpris.MediaPlayer2.TrackList` interface.
// https://specifications.freedesktop.org/mpris-spec/latest/TrackList_Interface.html
type TrackList struct {
	*MediaController
}

// URI is an unique resource identifier.
// https://specifications.freedesktop.org/mpris-spec/latest/Track_List_Interface.html#Simple-Type:Uri
type URI string

// MetadataMap is a mapping from metadata attribute names to values.
// https://specifications.freedesktop.org/mpris-spec/latest/Track_List_Interface.html#Mapping:Metadata_Map
type MetadataMap map[string]interface{}

func (m *MetadataMap) nonEmptyString(field, value string) {
	if value != "" {
		(*m)[field] = value
	}
}

func (m *MetadataMap) nonEmptySlice(field string, values []string) {
	toAdd := []string{}
	for _, v := range values {
		if v != "" {
			toAdd = append(toAdd, v)
		}
	}
	if len(toAdd) > 0 {
		(*m)[field] = toAdd
	}
}

// mapFromStatus returns a MetadataMap from the Song struct in mpd.
func mapFromStatus(s controller.Status) MetadataMap {
	if s.Song == nil {
		// No song
		return MetadataMap{
			"mpris:trackid": dbus.ObjectPath(basePath + "/TrackList/NoTrack"),
		}
	}

	m := &MetadataMap{
		"mpris:trackid": dbus.ObjectPath(fmt.Sprintf(TrackIDFormat, s.Song.Id)),
		"mpris:length":  s.Song.Duration * 1000 * 1000,
	}

	if s.Album != nil {
		m.nonEmptyString("xesam:album", s.Album.Name)
		m.nonEmptyString("mpris:artUrl", s.AlbumImageUrl)
	}
	if s.Artist != nil {
		m.nonEmptyString("xesam:artist", s.Artist.Name)
	}

	m.nonEmptyString("xesam:title", s.Song.Name)

	// mpris:artUrl, xesam:artist, xesam:url
	(*m)["xesam:trackNumber"] = s.Song.Index

	return *m
}
