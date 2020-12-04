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
	"fmt"
	"testing"
)

func Test_progressBar_Draw(t *testing.T) {
	type args struct {
		maximumValue int
		currentValue int
		width        int
	}
	tests := []struct {
		name   string
		fields args
		args   args
		want   string
	}{
		{
			args: args{
				maximumValue: 100,
				currentValue: 0,
				width:        10,
			},
			want: "┫╍╍╍╍╍╍╍╍╍╍┣",
		},
		{
			args: args{
				maximumValue: 100,
				currentValue: 25,
				width:        10,
			},
			want: "┫██▌╍╍╍╍╍╍╍┣",
		},
		{
			args: args{
				maximumValue: 30,
				currentValue: 16,
				width:        12,
			},
			want: "┫██████▎╍╍╍╍╍┣",
		},
		{
			args: args{
				maximumValue: 100,
				currentValue: 49,
				width:        10,
			},
			want: "┫████▊╍╍╍╍╍┣",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgressBar(tt.args.width, tt.args.maximumValue)
			if got := p.Draw(tt.args.currentValue); got != tt.want {
				t.Errorf("Draw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProgressBar_Draw(t *testing.T) {
	points := 20
	width := 20
	want := `┫╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍┣: 0/20
┫█╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍┣: 1/20
┫██╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍┣: 2/20
┫███╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍┣: 3/20
┫████╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍┣: 4/20
┫█████╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍┣: 5/20
┫██████╍╍╍╍╍╍╍╍╍╍╍╍╍╍┣: 6/20
┫███████╍╍╍╍╍╍╍╍╍╍╍╍╍┣: 7/20
┫████████╍╍╍╍╍╍╍╍╍╍╍╍┣: 8/20
┫█████████╍╍╍╍╍╍╍╍╍╍╍┣: 9/20
┫██████████╍╍╍╍╍╍╍╍╍╍┣: 10/20
┫███████████╍╍╍╍╍╍╍╍╍┣: 11/20
┫████████████╍╍╍╍╍╍╍╍┣: 12/20
┫█████████████╍╍╍╍╍╍╍┣: 13/20
┫██████████████╍╍╍╍╍╍┣: 14/20
┫███████████████╍╍╍╍╍┣: 15/20
┫████████████████╍╍╍╍┣: 16/20
┫█████████████████╍╍╍┣: 17/20
┫██████████████████╍╍┣: 18/20
┫███████████████████╍┣: 19/20
┫████████████████████┣: 20/20
`
	text := ""
	p := NewProgressBar(width, points)
	for i := 0; i < points+1; i++ {
		text += fmt.Sprintf("%s: %d/%d\n", p.Draw(i), i, points)
	}

	if text != want {
		t.Error("Invalid progress bar")
	}

}
