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

package widgets

import (
	"testing"
	"tryffel.net/go/jellycli/models"
)

func Test_albumSong_setText(t *testing.T) {
	type fields struct {
		showDiscNum   bool
		overrideIndex int
		width         int
		song          *models.Song
		index         int
	}
	tests := []struct {
		name   string
		fields fields

		// widget width

		wantDescription string
	}{
		{
			name: "one disc, one artist",
			fields: fields{
				song: &models.Song{
					Id:          "id",
					Name:        "A test song",
					Duration:    181,
					Index:       3,
					Album:       "An album",
					DiscNumber:  1,
					Artists:     nil,
					AlbumArtist: "Artist",
				},
				showDiscNum:   false,
				overrideIndex: -1,
				index:         0,
				width:         23,
			},
			wantDescription: "3. A test song   3:01\n",
		},
		{
			name: "one disc, one artist, override index",
			fields: fields{
				song: &models.Song{
					Id:          "id",
					Name:        "A test song",
					Duration:    181,
					Index:       3,
					Album:       "An album",
					DiscNumber:  1,
					Artists:     nil,
					AlbumArtist: "Artist",
				},
				showDiscNum:   false,
				overrideIndex: 4,
				index:         0,
				width:         23,
			},
			wantDescription: "4. A test song   3:01\n",
		},
		{
			name: "two discs, one artist",
			fields: fields{
				song: &models.Song{
					Id:          "id",
					Name:        "A test song",
					Duration:    181,
					Index:       3,
					Album:       "An album",
					DiscNumber:  2,
					Artists:     nil,
					AlbumArtist: "Artist",
				},
				showDiscNum:   true,
				overrideIndex: -1,
				index:         0,
				width:         23,
			},
			wantDescription: "2 3. A test song 3:01\n",
		},
		{
			name: "multiple artists",
			fields: fields{
				song: &models.Song{
					Id:          "id",
					Name:        "A test song",
					Duration:    181,
					Index:       3,
					Album:       "An album",
					DiscNumber:  2,
					Artists:     []models.IdName{{"", "Artist b"}, {"", "Artist c"}},
					AlbumArtist: "Artist",
				},
				showDiscNum:   true,
				overrideIndex: -1,
				index:         0,
				width:         30,
			},
			wantDescription: "2 3. A test song        3:01\n      Artist b, Artist c\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAlbumSong(tt.fields.song, tt.fields.showDiscNum, tt.fields.overrideIndex)
			a.SetRect(1, 1, tt.fields.width, 3)
			text := a.TextView.GetText(false)
			if tt.wantDescription != text {
				t.Errorf("format album song, want %s, got %s", tt.wantDescription, text)

			}
		})
	}
}
