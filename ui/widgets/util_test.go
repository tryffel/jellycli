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

package widgets

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
