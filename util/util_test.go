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

package util

import (
	"github.com/gdamore/tcell"
	"testing"
)

func TestPackKeyBindingName(t *testing.T) {
	tests := []struct {
		name      string
		key       tcell.Key
		maxLength int
		want      string
	}{
		{
			key:       tcell.KeyF6,
			maxLength: 0,
			want:      "F6",
		},
		{
			key:       tcell.KeyCtrlK,
			maxLength: 0,
			want:      "Ctrl-K",
		},
		{
			key:       tcell.KeyCtrlK,
			maxLength: 5,
			want:      "C-K",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PackKeyBindingName(tt.key, tt.maxLength); got != tt.want {
				t.Errorf("PackKeyBindingName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecToStringLong(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		{
			seconds: 600,
			want:    "10 mins",
		},
		{
			seconds: 150,
			want:    "2 mins",
		},
		{
			seconds: 60,
			want:    "1 min",
		},
		{
			seconds: 1260,
			want:    "21 mins",
		},
		{
			seconds: 3965,
			want:    "1 hour 6 mins",
		},
		{
			seconds: 7260,
			want:    "2 hours 1 min",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SecToStringApproximate(tt.seconds); got != tt.want {
				t.Errorf("SecToStringApproximate() = %v, want %v", got, tt.want)
			}
		})
	}
}
